function boot() {
  initWS()

  recolor()
  setInterval(recolor, 600)
}

function initWS() {
  if (!window.WebSocket) { return }
  var tempTable = document.querySelector('.js-temp-table')
  if (!tempTable) { return }
  startWS(buildWSURL(), tempTable)
}

function buildWSURL() {
  var host = location.host
  var protocol = location.protocol.replace("http", "ws")
  return `${protocol}//${host}/live`
}

var restartTO = null
var ws = null
function startWS(wsURL, tempTable) {
  console.log("opening " + wsURL)
  ws = new WebSocket(wsURL)
  ws.onmessage = function(event) {
    console.log(["WS UPDATE", event])
    tempTable.innerHTML = event.data
    recolor()
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

const HOUR = 3600

function recolor() {
  var now = new Date().getTime() / 1000
  var dates = document.querySelectorAll(".temp-date")
  for (var e of dates) {
    var ts = e.dataset.ts
    var age = now - ts
    var pe = e.parentElement
    if (age < HOUR) {
      pe.className = "age-fresh"
    } else if (age < 4*HOUR) {
      pe.className = "age-old"
    } else {
      pe.className = "age-expired"
    }
  }
}

boot()
