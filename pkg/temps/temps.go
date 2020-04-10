package temps

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"

	"github.com/spraints/temps/pkg/wu"
)

const temperatureUpdateInterval = 10 * time.Minute

type Temps struct {
	secret string
	wu     *wu.Client

	outdoorTemp fahrenheit
	sensors     sensorSlice
	lock        sync.RWMutex
}

type fahrenheit float64

type sensor struct {
	id          string
	Name        string
	Temperature fahrenheit
}

func New(config Config) *Temps {
	return &Temps{
		secret: config.Secret,
		wu:     wu.New(config.WundergroundAPIKey, config.WundergroundStationID),
	}
}

func (t *Temps) Register(mux chi.Router) {
	mux.Get("/", t.showTemps)
	mux.Put("/mytaglist/{secret}/{uuid}", t.handleTagData)
}

func (t *Temps) Poll(ctx context.Context) {
	ticker := time.NewTicker(temperatureUpdateInterval)
	defer ticker.Stop()

	for {
		if conditions, err := t.wu.GetCurrentConditions(ctx); err != nil {
			log.Print(err)
		} else {
			log.Printf("OUTDOORS -> %.0f F", conditions.ImperialTemperature)
			t.setOutdoorTemp(conditions)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (t *Temps) showTemps(w http.ResponseWriter, r *http.Request) {
	renderer := renderShowTemplateFahrenheit
	if strings.HasPrefix(r.Header.Get("User-Agent"), "curl") {
		renderer = renderShowText
	}

	sensors, outdoor := t.getDataForShow()
	if err := renderer(w, sensors, outdoor); err != nil {
		http.Error(w, "Error", 500)
		log.Printf("error rendering temperatures: %v", err)
	}
}

func renderShowText(w io.Writer, sensors []sensor, outdoor fahrenheit) error {
	show := func(label string, temp fahrenheit) {
		fmt.Fprintf(w, "%15s %.0f °F\n", label, temp)
	}

	show("Outside", outdoor)
	for _, sensor := range sensors {
		show(sensor.Name, sensor.Temperature)
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

	log.Printf("[%s] (%s) -> %.0f F", id, name, temperature)
	t.updateTagData(id, name, fahrenheit(temperature))
}

func (t *Temps) updateTagData(id string, name string, temperature fahrenheit) {
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

	t.outdoorTemp = fahrenheit(conditions.ImperialTemperature)
}

func (t *Temps) getDataForShow() ([]sensor, fahrenheit) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	sensors := make([]sensor, 0, len(t.sensors))
	for _, sensor := range t.sensors {
		sensors = append(sensors, *sensor)
	}

	return sensors, t.outdoorTemp
}
