package temps

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"

	"github.com/spraints/temps/pkg/templates"
	"github.com/spraints/temps/pkg/types"
	"github.com/spraints/temps/pkg/wu"
)

const (
	temperatureUpdateInterval = 10 * time.Minute
	outsideID                 = "outside"
)

type Temps struct {
	secret string
	now    func() time.Time

	weather   WeatherClient
	store     Store
	templates TemplateLoader

	ws wsData
}

type WeatherClient interface {
	GetCurrentConditions(ctx context.Context) (*wu.Conditions, error)
}

type Store interface {
	All() ([]types.Measurement, error)
	Put(types.Measurement) error
}

type TemplateLoader interface {
	Get(path string) templates.Template
}

func New(weather WeatherClient, store Store, templates TemplateLoader, opts ...Option) *Temps {
	t := &Temps{}
	t.weather = weather
	t.store = store
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
			if err := t.store.Put(types.Measurement{
				ID:          outsideID,
				Name:        "Outdoors",
				Temperature: conditions.Temperature,
				MeasuredAt:  t.now(),
			}); err != nil {
				log.Printf("error: %v", err)
			} else {
				t.updateWSTemps()
			}
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
	Temps []temp `json:"temps"`
}

type temp struct {
	Label       string            `json:"location"`
	Temperature types.Temperature `json:"value"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (t *Temps) showTemps(w http.ResponseWriter, r *http.Request) {
	data := &showData{
		Temps: t.getDataForShow(),
	}
	tmpl := "show.html.tmpl"
	if strings.HasPrefix(r.Header.Get("User-Agent"), "curl") || strings.Contains(r.Header.Get("Accept"), "text/plain") {
		tmpl = "show.text.tmpl"
		w.Header().Set("Content-Type", "text/plain")
	}
	if strings.Contains(r.Header.Get("Accept"), "json") {
		tmpl = "show.text.tmpl"
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Error", 500)
			log.Printf("error rendering json temperatures: %v", err)
			return
		}
	}
	if err := t.templates.Get(tmpl).Execute(w, data); err != nil {
		http.Error(w, "Error", 500)
		log.Printf("error rendering temperatures: %v", err)
	}
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
	if err := t.store.Put(types.Measurement{
		ID:          "wirelesstag-" + id,
		Name:        name,
		Temperature: types.Celsius(temperature),
		MeasuredAt:  t.now(),
	}); err != nil {
		log.Printf("error: %v", err)
	} else {
		t.updateWSTemps()
	}
}

func (t *Temps) getDataForShow() []temp {
	m, err := t.store.All()
	if err != nil {
		log.Printf("error getting values: %v", err)
		return nil
	}
	n := sensorSlice(m)
	sort.Sort(n)

	temps := make([]temp, 0, len(n))
	for _, sensor := range n {
		log.Printf(">> %q", sensor.ID)
		temps = append(temps, temp{
			Label:       sensor.Name,
			Temperature: sensor.Temperature,
			UpdatedAt:   sensor.MeasuredAt,
		})
	}

	return temps
}
