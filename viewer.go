package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// A connection from a viewer
type Viewer struct {
	Websocket        *websocket.Conn // the websocket connection
	OutboundMessages chan string     // buffered channel of outbound messages
	Unregister       chan *Viewer    // unregister channel of the hub this connection is subsribed to, to handle client closing connection
}

// reads messages from the channel and writes them to the websocket
func (v *Viewer) writeUntilClose() {
	for message := range v.OutboundMessages {
		if err := v.Websocket.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("forcing unregister:", err)
			v.Unregister <- v
			break
		}
	}
}
