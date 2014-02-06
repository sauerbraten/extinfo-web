package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net"
	"sync"
)

// A connection from a viewer
type Connection struct {
	Websocket               *websocket.Conn  // the websocket connection
	OutboundBasicUpdates    chan string      // buffered channel of outbound basic update messages; closing this channel makes the WS connection close, too
	OutboundExtendedUpdates chan string      // buffered channel of outbound extended update messages
	Unregister              chan *Connection // unregister channel of the hub this connection is subsribed to, to handle client closing connection
}

func newBasicConnection(ws *websocket.Conn) *Connection {
	return &Connection{
		Websocket:               ws,
		OutboundBasicUpdates:    make(chan string),
		OutboundExtendedUpdates: nil,
	}
}

func newExtendedConnection(ws *websocket.Conn) *Connection {
	c := newBasicConnection(ws)
	c.OutboundExtendedUpdates = make(chan string)
	return c
}

func (c *Connection) readUntilClose(wg sync.WaitGroup) {
	var message string
	for {
		// receive message
		if err := websocket.Message.Receive(c.Websocket, &message); err != nil {
			// client closed the connection
			log.Println("forcing unregister", err)
			c.Unregister <- c
			break
		}
		c.processMessage(message)
	}
	wg.Done()
}

// Parses the message coming in from the websocket. Incoming messages are of the form "abc.com:1234" and mean that the client wants to subscribe to a server poller.
func (c *Connection) processMessage(message string) {
	// get address
	addr, err := net.ResolveUDPAddr("udp4", message)
	if err != nil {
		log.Println(err)
		close(c.OutboundBasicUpdates)
		return
	}

	h, ok := hubs[addr.String()]

	if !ok {
		// get hostname
		names, err := net.LookupAddr(addr.IP.String())
		if err != nil {
			log.Println(err)
			close(c.OutboundBasicUpdates)
			return
		}

		hostname := ""

		// use first hostname found
		if len(names) > 0 {
			hostname = names[0]
			// cut off trailing '.'
			hostname = hostname[:len(hostname)-1]
		}

		// spawn new poller and hub for new sauer server
		h, err = newHubWithPoller(addr, hostname)
		if err != nil {
			log.Println(err)
			close(c.OutboundBasicUpdates)
			return
		}

		hubs[addr.String()] = h

		go h.run()

		log.Println("spawned new hub for", addr.String())
	}

	c.Unregister = h.Unregister
	h.Register <- c
	h.Poller.getAllInfoOnce()
}

// reads messages from the two update channels and writes them to the websocket
func (c *Connection) writeUpdatesUntilClose(wg sync.WaitGroup) {
	var err error
	for {
		select {
		case message, ok := <-c.OutboundBasicUpdates:
			if ok {
				err = websocket.Message.Send(c.Websocket, message)
			}
			if !ok || err != nil {
				log.Println("forcing unregister:", err)
				c.Unregister <- c
				break
			}

		case message, ok := <-c.OutboundExtendedUpdates:
			if ok {
				err = websocket.Message.Send(c.Websocket, message)
			}
			if !ok || err != nil {
				log.Println("forcing unregister:", err)
				c.Unregister <- c
				break
			}
		}
	}
	wg.Done()
}

// registers websockets
func basicUpdatesWSHandler(ws *websocket.Conn) {
	registerViewer(newBasicConnection(ws))
}

func extendedUpdatesWSHandler(ws *websocket.Conn) {
	registerViewer(newExtendedConnection(ws))
}

func registerViewer(c *Connection) {
	var wg sync.WaitGroup
	wg.Add(2)

	go c.writeUpdatesUntilClose(wg)
	go c.readUntilClose(wg)

	wg.Wait()
	c.Websocket.Close()
}
