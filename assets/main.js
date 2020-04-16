import {BroadcastChannel, createLeaderElection} from 'broadcast-channel'

function initWS(tempTable, updates) {
  if (!window.WebSocket) { return }
  startWS(buildWSURL(), tempTable, updates)
}

function buildWSURL() {
  var host = location.host
  var protocol = location.protocol.replace("http", "ws")
  return `${protocol}//${host}/live`
}

var restartTO = null
var ws = null
function startWS(wsURL, tempTable, updates) {
  console.log("opening " + wsURL)
  ws = new WebSocket(wsURL)
  ws.onmessage = function(event) {
    console.log("update received from websocket")
    tempTable.innerHTML = event.data
    updates.postMessage(event.data)
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

var tempTable = document.querySelector('.js-temp-table')
if (tempTable) {
  const tmtyl = new BroadcastChannel('tmtyl')
  const updates = new BroadcastChannel('temps')

  const elector = createLeaderElection(tmtyl)
  elector.awaitLeadership().then(() => { initWS(tempTable, updates) })

  updates.onmessage = function(msg) {
    console.log("update received from leader")
    tempTable.innerHTML = msg
  }
}
