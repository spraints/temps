package temps

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/spraints/temps/pkg/wu"
)

type Temps struct {
	secret string
	wu     *wu.Client

	outdoorTemp fahrenheit
	sensors     []*sensor
	lock        sync.RWMutex
}

type fahrenheit float32

type sensor struct {
	Name        string
	Temperature fahrenheit
}

func New(config Config) *Temps {
	return &Temps{
		secret: config.Secret,
		wu:     wu.New(config.WundergroundAPIKey, config.WundergroundStationID),
	}
}

func (t *Temps) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", t.showTemps)
	mux.HandleFunc("/mytaglist/"+t.secret+"/", t.updateTagTemp)
}

func (t *Temps) Poll(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		if conditions, err := t.wu.GetCurrentConditions(ctx); err != nil {
			log.Print(err)
		} else {
			t.setOutdoorTemp(conditions)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (t *Temps) showTemps(w http.ResponseWriter, _ *http.Request) {
	sensors, outdoor := t.getDataForShow()

	if err := renderShowTemplateFahrenheit(w, sensors, outdoor); err != nil {
		http.Error(w, "Error", 500)
		log.Printf("error rendering temperatures: %v", err)
	}
}

func (t *Temps) updateTagTemp(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO", 500)
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
