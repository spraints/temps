package temps

import (
	"context"
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

	"github.com/spraints/temps/pkg/templates"
	"github.com/spraints/temps/pkg/units"
	"github.com/spraints/temps/pkg/wu"
)

const temperatureUpdateInterval = 10 * time.Minute

type Temps struct {
	secret string
	now    func() time.Time

	weather   WeatherClient
	templates TemplateLoader

	lock          sync.RWMutex
	outdoorTemp   units.Temperature
	outdoorTempAt time.Time
	sensors       sensorSlice

	ws wsData
}

type WeatherClient interface {
	GetCurrentConditions(ctx context.Context) (*wu.Conditions, error)
}

type TemplateLoader interface {
	Get(path string) templates.Template
}

type sensor struct {
	id          string
	Name        string
	Temperature units.Temperature
	UpdatedAt   time.Time
}

func New(weather WeatherClient, templates TemplateLoader, opts ...Option) *Temps {
	t := &Temps{}
	t.outdoorTemp = units.Fahrenheit(0)
	t.weather = weather
	t.templates = templates
	t.now = time.Now
	for _, opt := range opts {
		opt(t)
	}
	t.updateWSTemps()
	return t
}

func (t *Temps) Register(mux chi.Router) {
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
		return
	}
	t.runWS(c)
}

type showData struct {
	Temps []temp
}

type temp struct {
	Label       string
	Temperature units.Temperature
	UpdatedAt   time.Time
}

func (t *Temps) showTemps(w http.ResponseWriter, r *http.Request) {
	data := &showData{
		Temps: t.getDataForShow(),
	}
	tmpl := "show.html.tmpl"
	if strings.HasPrefix(r.Header.Get("User-Agent"), "curl") {
		tmpl = "show.text.tmpl"
	}
	if err := t.templates.Get(tmpl).Execute(w, data); err != nil {
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
			sensor.UpdatedAt = t.now()
			return
		}
	}

	sensor := &sensor{
		id:          id,
		Name:        name,
		Temperature: temperature,
		UpdatedAt:   t.now(),
	}

	t.sensors = append(t.sensors, sensor)
	sort.Sort(t.sensors)
}

func (t *Temps) setOutdoorTemp(conditions *wu.Conditions) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.outdoorTemp = conditions.Temperature
	t.outdoorTempAt = t.now()
}

func (t *Temps) getDataForShow() []temp {
	t.lock.RLock()
	defer t.lock.RUnlock()

	temps := make([]temp, 0, 1+len(t.sensors))
	temps = append(temps, temp{Label: "Outside", Temperature: t.outdoorTemp, UpdatedAt: t.outdoorTempAt})

	for _, sensor := range t.sensors {
		temps = append(temps, temp{Label: sensor.Name, Temperature: sensor.Temperature, UpdatedAt: sensor.UpdatedAt})
	}

	return temps
}
