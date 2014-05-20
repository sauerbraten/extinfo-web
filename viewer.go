package main

import (
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
	"sync"
)

// A connection from a viewer
type Viewer struct {
	Websocket        *websocket.Conn // the websocket connection
	OutboundMessages chan string     // buffered channel of outbound messages
	Unregister       chan *Viewer    // unregister channel of the hub this connection is subsribed to, to handle client closing connection
}

func (v *Viewer) readUntilClose(wg *sync.WaitGroup) {
	for {
		// receive message
		messageType, payload, err := v.Websocket.ReadMessage()
		if err != nil {
			if err != io.EOF {
				// client did no cleanly close the connection
				log.Println("forcing unregister:", err)
			}

			v.Unregister <- v
			break
		}

		if messageType != websocket.TextMessage {
			// ignore everything but text messages
			continue
		}

		message := string(payload)

		err = v.processMessage(message)
		if err != nil {
			// message was bad
			log.Println("bad message, closing websocket:", err)
			v.OutboundMessages <- "error\tno server at " + message + "\ttry a different address and/or port"
			close(v.OutboundMessages)
			break
		}
	}
	wg.Done()
}

// Parses the message coming in from the websocket. Incoming messages are of the form "abc.com:1234" and mean that the client wants to subscribe to a server poller.
func (v *Viewer) processMessage(message string) error {
	// get address
	addr, err := net.ResolveUDPAddr("udp4", message)

	if err != nil {
		return err
	}

	h, ok := hubs[addr.String()]

	if !ok {
		// get hostname
		hostname := addr.IP.String()

		names, err := net.LookupAddr(addr.IP.String())
		if err == nil && len(names) > 0 {
			// use first hostname found
			hostname = names[0]
			// cut off trailing '.'
			hostname = hostname[:len(hostname)-1]
		}

		// spawn new poller and hub for new sauer server
		newHub, err := newHubWithPoller(addr, hostname)
		if err != nil {
			return err
		}

		hubs[addr.String()] = newHub
		h = newHub

		go newHub.run()

		log.Println("spawned new hub for", addr.String())
	}

	v.Unregister = h.Unregister
	h.Register <- v
	h.Poller.getAllOnce()

	return err
}

// reads messages from the channel and writes them to the websocket
func (v *Viewer) writeUntilClose(wg *sync.WaitGroup) {
	for message := range v.OutboundMessages {
		if err := v.Websocket.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("forcing unregister:", err)
			v.Unregister <- v
			break
		}
	}
	wg.Done()
}
