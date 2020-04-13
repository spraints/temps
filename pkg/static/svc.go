package static

import (
	"net/http"

	"github.com/go-chi/chi"
)

var Svc Static

type Static struct{}

func (Static) Register(mux chi.Router) {
	static := func(path string, contentType string, content []byte) {
		mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			w.Write(content)
		})
	}
	static(AppJS, "text/javascript", []byte(appJS))
}
