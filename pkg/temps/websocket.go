package temps

import (
	"bytes"
	"log"

	"github.com/gorilla/websocket"
)

func (t *Temps) runWS(ws *websocket.Conn) {
	defer ws.Close()
	if err := t.sendWSTemps(ws); err != nil {
		log.Printf("unable to send temps on websocket start: %v", err)
		return
	}

	t.wsCond.L.Lock()
	defer t.wsCond.L.Unlock()
	for {
		t.wsCond.Wait()
		if err := t.sendWSTemps(ws); err != nil {
			log.Printf("unable to send temp update: %v", err)
			return
		}
	}
}

func (t *Temps) sendWSTemps(ws *websocket.Conn) error {
	return ws.WriteMessage(websocket.TextMessage, t.tempTable)
}

func (t *Temps) updateWSTemps() {
	t.wsCond.L.Lock()
	defer t.wsCond.L.Unlock()

	var buf bytes.Buffer
	if err := showFrag(&buf, t.getDataForShow()); err != nil {
		log.Printf("error updating table for websockets: %v", err)
		return
	}

	t.tempTable = buf.Bytes()
	t.wsCond.Broadcast()
}
