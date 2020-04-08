package temps

import (
	"context"
	"log"
	"net/http"
	"sync"
)

type Temps struct {
	secret      string
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
	return &Temps{secret: config.Secret}
}

func (t *Temps) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", t.showTemps)
	mux.HandleFunc("/mytaglist/"+t.secret+"/", t.updateTagTemp)
}

func (t *Temps) Poll(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Printf("TODO - actually poll!")
		return
	}
}

func (t *Temps) showTemps(w http.ResponseWriter, _ *http.Request) {
	sensors, outdoor := t.getDataForShow()

	if err := renderShowTemplateFahrenheit(w, sensors, outdoor); err != nil {
		http.Error(w, "Error", 500)
		log.Printf("error rendering temperatures: %v", err)
	}
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

func (t *Temps) updateTagTemp(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO", 500)
}
