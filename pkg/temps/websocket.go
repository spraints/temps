package temps

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const wsUpdateInterval = 10 * time.Second

type wsData struct {
	lock      sync.RWMutex
	serial    int
	tempTable []byte
}

func (t *Temps) runWS(ws *websocket.Conn) {
	log.Printf("websocket accepted %v", ws.RemoteAddr())

	defer ws.Close()

	tick := time.NewTicker(wsUpdateInterval)
	defer tick.Stop()

	lastSent := 0
	for range tick.C {
		t.ws.lock.RLock()
		if t.ws.serial > lastSent {
			lastSent = t.ws.serial
			if err := ws.WriteMessage(websocket.TextMessage, t.ws.tempTable); err != nil {
				log.Printf("error sending temps to websocket %v: %v", ws.RemoteAddr(), err)
				t.ws.lock.RUnlock()
				return
			}
		}
		t.ws.lock.RUnlock()
	}
}

func (t *Temps) updateWSTemps() {
	t.ws.lock.Lock()
	defer t.ws.lock.Unlock()

	var buf bytes.Buffer
	if err := showFrag(&buf, t.getDataForShow()); err != nil {
		log.Printf("error updating table for websockets: %v", err)
		return
	}

	t.ws.serial = t.ws.serial + 1
	t.ws.tempTable = buf.Bytes()
}
