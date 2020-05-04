package temps

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/spraints/temps/pkg/types"
)

// nginx times out connections after 60s, so ping more frequently than that.
const wsTimeout = 55 * time.Second

type wsUpdateMessage []byte

type wsData struct {
	subscriberChange chan subchange
	updates          chan types.Measurement

	subscribers map[*websocket.Conn]chan wsUpdateMessage
}

type subchange struct {
	add  bool
	conn *websocket.Conn
}

func (t *Temps) initWS() {
	t.ws.subscriberChange = make(chan subchange)
	t.ws.updates = make(chan types.Measurement, 10)
	t.ws.subscribers = make(map[*websocket.Conn]chan wsUpdateMessage)
	go t.pumpWS()
}

func (t *Temps) runWS(conn *websocket.Conn) {
	remote := conn.RemoteAddr()
	log.Printf("[%v] websocket accepted", remote)

	t.ws.subscriberChange <- subchange{add: true, conn: conn}

	conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("[%v] websocket closed: %d %s", remote, code, text)
		t.ws.subscriberChange <- subchange{add: false, conn: conn}
		return nil
	})
}

func (t *Temps) updateWSTemps(meas types.Measurement) {
	t.ws.updates <- meas
}

func (t *Temps) pumpWS() {
	for {
	tick:
		select {
		case change := <-t.ws.subscriberChange:
			if change.add {
				conn := change.conn
				t.ws.subscribers[conn] = make(chan wsUpdateMessage, 2)
				if html, err := t.renderWSUpdate(); err != nil {
					log.Printf("error updating table for websocket init: %v", err)
				} else {
					t.ws.subscribers[conn] <- html
				}
				go sendUpdates(conn, t.ws.subscribers[conn], func() {
					log.Printf("[%v] closing websocket after too many errors", conn.RemoteAddr())
					t.ws.subscriberChange <- subchange{add: false, conn: conn}
				})
			} else {
				if c, ok := t.ws.subscribers[change.conn]; ok {
					close(c)
					delete(t.ws.subscribers, change.conn)
				}
			}

		case <-t.ws.updates:
			if len(t.ws.subscribers) == 0 {
				break tick
			}

			// flush all other pending updates
		flush:
			for {
				select {
				case <-t.ws.updates:
				default:
					break flush
				}
			}

			html, err := t.renderWSUpdate()
			if err != nil {
				log.Printf("error updating table for websockets: %v", err)
				break tick
			}
			log.Printf("sending %d bytes to %d websockets", len(html), len(t.ws.subscribers))

			for _, ch := range t.ws.subscribers {
				select {
				case ch <- html:
				default:
				}
			}
		}
	}
}

func sendUpdates(conn *websocket.Conn, htmls <-chan wsUpdateMessage, circuitbreaker func()) {
	errCount := 0
	for {
		if errCount > 3 {
			circuitbreaker()
		}

		select {
		case html, ok := <-htmls:
			if !ok {
				// websocket is closed
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, html); err != nil {
				errCount += 1
				log.Printf("[%v/%d] error sending update to websocket: %v", conn.RemoteAddr(), errCount, err)
				break
			}
			errCount = 0
		case <-time.After(wsTimeout):
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				errCount += 1
				log.Printf("[%v/%d] error sending ping to websocket: %v", conn.RemoteAddr(), errCount, err)
				break
			}
			errCount = 0
		}
	}
}

func (t *Temps) renderWSUpdate() (wsUpdateMessage, error) {
	var buf bytes.Buffer
	if err := t.templates.Get("show.html.tmpl").ExecuteTemplate(&buf, "table", t.getDataForShow()); err != nil {
		return nil, err
	}
	return wsUpdateMessage(buf.Bytes()), nil
}
