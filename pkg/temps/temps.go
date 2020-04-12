package temps

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"

	"github.com/spraints/temps/pkg/units"
	"github.com/spraints/temps/pkg/wu"
)

const temperatureUpdateInterval = 10 * time.Minute

type Temps struct {
	secret  string
	weather WeatherClient

	outdoorTemp units.Temperature
	sensors     sensorSlice
	lock        sync.RWMutex

	wsCond    *sync.Cond
	tempTable []byte
}

type WeatherClient interface {
	GetCurrentConditions(ctx context.Context) (*wu.Conditions, error)
}

type sensor struct {
	id          string
	Name        string
	Temperature units.Temperature
}

func New(opts ...Option) *Temps {
	t := &Temps{}
	t.outdoorTemp = units.Fahrenheit(0)
	for _, opt := range opts {
		opt(t)
	}
	t.wsCond = sync.NewCond(&sync.Mutex{})
	t.updateWSTemps()
	return t
}

func (t *Temps) Register(mux chi.Router) {
	registerStatic(mux)
	mux.Get("/", t.showTemps)
	mux.Get("/live", t.live)
	if t.secret != "" {
		mux.Put("/mytaglist/{secret}/{uuid}", t.handleTagData)
	}
}

func (t *Temps) Poll(ctx context.Context) {
	ticker := time.NewTicker(temperatureUpdateInterval)
	defer ticker.Stop()

	for {
		if conditions, err := t.weather.GetCurrentConditions(ctx); err != nil {
			log.Print(err)
		} else {
			log.Printf("OUTDOORS -> %.2f F", conditions.Temperature)
			t.setOutdoorTemp(conditions)
			t.updateWSTemps()
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (t *Temps) live(w http.ResponseWriter, r *http.Request) {
	u := websocket.Upgrader{}
	c, err := u.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("couldn't upgrade websocket connection: %v", err)
		http.Error(w, "Error", 500)
		return
	}
	t.runWS(c)
}

func (t *Temps) showTemps(w http.ResponseWriter, r *http.Request) {
	renderer := func(w io.Writer, temps []temp) error { return showHTML(w, getWSURL(r), temps) }
	if strings.HasPrefix(r.Header.Get("User-Agent"), "curl") {
		renderer = showText
	}

	if err := renderer(w, t.getDataForShow()); err != nil {
		http.Error(w, "Error", 500)
		log.Printf("error rendering temperatures: %v", err)
	}
}

func getWSURL(r *http.Request) string {
	var wsURL url.URL
	wsURL.Scheme = "ws"
	wsURL.Host = r.Host
	wsURL.Path = "/live"
	return wsURL.String()
}

func showText(w io.Writer, temps []temp) error {
	for _, temp := range temps {
		if _, err := fmt.Fprintf(w, "%-15s %3.0f °F / %3.0f °C\n", temp.Label, temp.Temperature.Fahrenheit(), temp.Temperature.Celsius()); err != nil {
			return err
		}
	}

	return nil
}

func (t *Temps) handleTagData(w http.ResponseWriter, r *http.Request) {
	defer w.Write([]byte("OK!\n"))

	secret := chi.URLParam(r, "secret")
	id := chi.URLParam(r, "uuid")

	if secret != t.secret || id == "" {
		return
	}

	name := r.FormValue("name")
	temperature, err := strconv.ParseFloat(r.FormValue("temperature"), 64)
	if err != nil {
		log.Print(err)
		return
	}

	log.Printf("[%s] (%s) -> %.3f C", id, name, temperature)
	t.updateTagData(id, name, units.Celsius(temperature))
	t.updateWSTemps()
}

func (t *Temps) updateTagData(id string, name string, temperature units.Temperature) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, sensor := range t.sensors {
		if sensor.id == id {
			if name != "" {
				sensor.Name = name
			}
			sensor.Temperature = temperature
			return
		}
	}

	sensor := &sensor{
		id:          id,
		Name:        name,
		Temperature: temperature,
	}

	t.sensors = append(t.sensors, sensor)
	sort.Sort(t.sensors)
}

func (t *Temps) setOutdoorTemp(conditions *wu.Conditions) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.outdoorTemp = conditions.Temperature
}

func (t *Temps) getDataForShow() []temp {
	t.lock.RLock()
	defer t.lock.RUnlock()

	temps := make([]temp, 0, 1+len(t.sensors))
	temps = append(temps, temp{Label: "Outside", Temperature: t.outdoorTemp})

	for _, sensor := range t.sensors {
		temps = append(temps, temp{Label: sensor.Name, Temperature: sensor.Temperature})
	}

	return temps
}
