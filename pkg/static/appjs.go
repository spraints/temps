package static

const AppJS = "/app.js"

const appJS = `
function initWS() {
  if (!window.WebSocket) { return }
  var tempTable = document.querySelector('.js-temp-table')
  if (!tempTable) { return }
  var wsURL = tempTable.getAttribute('data-ws-url')
  if (!wsURL) { return }
  startWS(wsURL, tempTable)
}

var restartTO = null
var ws = null
function startWS(wsURL, tempTable) {
  console.log("opening " + wsURL)
  ws = new WebSocket(wsURL)
  ws.onmessage = function(event) {
    console.log(["WS UPDATE", event])
    tempTable.innerHTML = event.data
  }
  var restart = function() {
    ws.close()
    if (!restartTO) {
      console.log("restarting websocket in 10s...")
      restartTO = setTimeout(function() {
        restartTO = null
        startWS(wsURL, tempTable)
      }, 10000)
    }
  }
  ws.onerror = restart
  ws.onclose = restart
}

initWS()
`
