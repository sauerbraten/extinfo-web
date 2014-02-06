package main

import (
	"code.google.com/p/go.net/websocket"
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
	var message string
	for {
		// receive message
		err := websocket.Message.Receive(v.Websocket, &message)
		if err != nil {
			// client closed the connection
			log.Println("forcing unregister:", err)
			v.Unregister <- v
			break
		}

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
		names, err := net.LookupAddr(addr.IP.String())
		if err != nil {
			return err
		}

		hostname := ""

		// use first hostname found
		if len(names) > 0 {
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
		if err := websocket.Message.Send(v.Websocket, message); err != nil {
			log.Println("forcing unregister:", err)
			v.Unregister <- v
			break
		}
	}
	wg.Done()
}

// registers websockets
func websocketHandler(ws *websocket.Conn) {
	v := &Viewer{
		Websocket:        ws,
		OutboundMessages: make(chan string),
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go v.writeUntilClose(wg)
	go v.readUntilClose(wg)

	wg.Wait()
	v.Websocket.Close()
}
