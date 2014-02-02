package main

import (
	"log"
	"net"
	"time"
)

type Hub struct {
	Connections map[*Connection]bool // registered connections
	Poller      *Poller
	//Updates     chan string      // updates to send to subscribers
	//Quit        chan bool        // channel for hub and poller to notify each other to stop
	Register   chan *Connection // register requests from connections
	Unregister chan *Connection // unregister requests from connections
	Address    *net.UDPAddr     // address of the game server this hub is for
	Hostname   string           // the game server's hostname
}

func newHubWithPoller(addr *net.UDPAddr, hostname string) (h *Hub, err error) {
	poller, err := newPoller(addr)
	if err != nil {
		return
	}

	h = &Hub{
		Connections: map[*Connection]bool{},
		Poller:      poller,
		Register:    make(chan *Connection),
		Unregister:  make(chan *Connection),
		Address:     addr,
		Hostname:    hostname,
	}

	return
}

func (h *Hub) run() {
	go h.Poller.pollForever()
	for {
		select {
		case conn := <-h.Register:
			h.Connections[conn] = true
			log.Println("websocket registered at hub", h.Address.String())

		case conn := <-h.Unregister:
			delete(h.Connections, conn)
			log.Println("connections of", h.Address.String(), ":", h.Connections)
			log.Println("websocket unregistered from hub", h.Address.String())

			log.Println("len of connections:", len(h.Connections))

			// if no subscriber left
			if len(h.Connections) == 0 {
				// end poller
				h.Poller.Quit <- true

				log.Println("hubs:", hubs)

				log.Println("deleting hub")

				// remove hub
				delete(hubs, h.Address.String())

				log.Println("hubs:", hubs)

				// end goroutine
				return
			}

		case message := <-h.Poller.Updates:
			// concurrently send message to all subscribers with a 5 second timeout
			for conn := range h.Connections {
				go func(conn *Connection, message string) {
					select {
					case conn.OutboundMessages <- message:
					case <-time.After(5 * time.Second):
						log.Println("forcing unregister")
						h.Unregister <- conn
					}
				}(conn, message)
			}
		}
	}
}
