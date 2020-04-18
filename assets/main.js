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
  const chan = new BroadcastChannel('temps')
  window.chan = chan

  const elector = createLeaderElection(chan)
  elector.awaitLeadership().then(() => {
    document.title = document.title + " - LEADER w timeout"
    setTimeout(() => initWS(tempTable, chan), 10)
  })

  chan.onmessage = function(msg) {
    console.log("update received from leader")
    tempTable.innerHTML = msg
  }
}
