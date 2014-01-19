package main

import (
	"log"
	"time"
)

type server struct {
	IP   string
	Host string
	Port int
}

type hub struct {
	Connections map[*connection]bool // registered connections; exported for template embedding
	updates     chan string          // updates to send to the connections
	quit        chan bool            // channel to end polling
	register    chan *connection     // register requests from connections
	unregister  chan *connection     // unregister requests from connections
	S           server               // address and port of the server; exported for template embedding
}

func (h hub) run() {
	t := time.Tick(30 * time.Second)
	for {
		select {
		case c := <-h.register:
			h.Connections[c] = true
		case c := <-h.unregister:
			log.Println("websocket unregistered from hub", h.S.IP, "port", h.S.Port)
			delete(h.Connections, c)
			close(c.send)
		case m := <-h.updates:
			for c := range h.Connections {
				select {
				case c.send <- m:
				default:
					delete(h.Connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		case <-t:
			// on each tick, the hub checks if it still has websockets connected
			// if not, the hub stops the poller and deletes itself
			if len(h.Connections) == 0 {
				// end poller
				h.quit <- true

				// remove hub
				delete(hubs, h.S)

				// end goroutine
				return
			}
		}
	}
}
