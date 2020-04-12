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
	remote := ws.RemoteAddr()
	log.Printf("[%v] websocket accepted", remote)

	defer ws.Close()

	tick := time.NewTicker(wsUpdateInterval)
	defer tick.Stop()

	done := make(chan struct{})
	defer close(done)
	ws.SetCloseHandler(func(code int, text string) error {
		log.Printf("[%v] websocket closed: %d %s", remote, code, text)
		close(done)
		return nil
	})

	lastSent := 0
	for {
		t.ws.lock.RLock()
		if t.ws.serial > lastSent {
			log.Printf("[%v] update to ws.serial = %d", remote, t.ws.serial)
			lastSent = t.ws.serial
			if err := ws.WriteMessage(websocket.TextMessage, t.ws.tempTable); err != nil {
				log.Printf("[%v] error sending temps to websocket: %v", remote, err)
				t.ws.lock.RUnlock()
				return
			}
		}
		t.ws.lock.RUnlock()

		select {
		case <-tick.C:
			continue
		case <-done:
			return
		}
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
	log.Printf("ws.serial = %d", t.ws.serial)
}
