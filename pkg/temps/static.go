package temps

import (
	"net/http"

	"github.com/go-chi/chi"
)

var appJS = `
function init() {
  if (!window.WebSocket) { return }
  var tempTable = document.querySelector('.js-temp-table')
  if (!tempTable) { return }
  var wsURL = tempTable.getAttribute('data-ws-url')
  if (!wsURL) { return }
  var ws = new WebSocket(wsURL)
  ws.onmessage = function(event) {
    tempTable.innerHTML = event.data
  }
}

init()
`

func registerStatic(mux chi.Router) {
	static := func(path string, contentType string, content []byte) {
		mux.Get(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			w.Write(content)
		})
	}
	static("/app.js", "text/javascript", []byte(appJS))
}
