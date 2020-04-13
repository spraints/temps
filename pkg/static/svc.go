package static

import (
	"github.com/go-chi/chi"
)

var Svc Static

type Static struct{}

func (Static) Register(mux chi.Router) {
}
