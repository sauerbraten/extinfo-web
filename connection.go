package main

import (
	"code.google.com/p/go.net/websocket"
	"io"
	"log"
	"net"
	"sync"
)

// A connection from a viewer
type Connection struct {
	Websocket        *websocket.Conn  // the websocket connection
	OutboundMessages chan string      // buffered channel of outbound messages
	Unregister       chan *Connection // unregister channel of the hub this connection is subsribed to, to handle client closing connection
}

func (c *Connection) readUntilClose(wg sync.WaitGroup) {
	var message string
	for {
		// receive message
		if err := websocket.Message.Receive(c.Websocket, &message); err != nil {
			if err == io.EOF {
				// expected; client closed the connection
				c.Unregister <- c
				break
			} else {
				log.Println(err)
				continue
			}
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
		close(c.OutboundMessages)
		return
	}

	h, ok := hubs[addr.String()]

	if !ok {
		// get hostname
		names, err := net.LookupAddr(addr.IP.String())
		if err != nil {
			log.Println(err)
			close(c.OutboundMessages)
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
			close(c.OutboundMessages)
			return
		}

		hubs[addr.String()] = h

		go h.run()

		log.Println("spawned new hub for", addr.String())
	}

	c.Unregister = h.Unregister
	h.Register <- c
	h.Poller.getAllOnce()
}

// reads messages from the channel and writes them to the websocket
func (c *Connection) writeUntilClose(wg sync.WaitGroup) {
	for message := range c.OutboundMessages {
		err := websocket.Message.Send(c.Websocket, message)
		if err != nil {
			log.Println("sending:", err)
			break
		}
	}
	wg.Done()
}

// registers websockets
func websocketHandler(ws *websocket.Conn) {
	c := &Connection{
		Websocket:        ws,
		OutboundMessages: make(chan string),
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go c.writeUntilClose(wg)
	go c.readUntilClose(wg)

	wg.Wait()
	c.Websocket.Close()
}
