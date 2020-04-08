package temps

import (
	"context"
	"log"
	"net/http"
)

type Temps struct {
	secret string
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

func (t *Temps) showTemps(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO", 500)
}

func (t *Temps) updateTagTemp(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "TODO", 500)
}
