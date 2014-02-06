package main

import (
	"log"
	"net"
	"time"
)

type Hub struct {
	Subscribers map[*Viewer]bool // registered viewers
	Poller      *Poller
	Register    chan *Viewer // register requests from viewers
	Unregister  chan *Viewer // unregister requests from viewers
	Address     *net.UDPAddr // address of the game server this hub is for
	Hostname    string       // the game server's hostname
}

func newHubWithPoller(addr *net.UDPAddr, hostname string) (h *Hub, err error) {
	var poller *Poller
	poller, err = newPoller(addr)
	if err != nil {
		return
	}

	h = &Hub{
		Subscribers: map[*Viewer]bool{},
		Poller:      poller,
		Register:    make(chan *Viewer),
		Unregister:  make(chan *Viewer),
		Address:     addr,
		Hostname:    hostname,
	}

	return
}

func (h *Hub) run() {
	go h.Poller.pollForever()
	for {
		select {
		case viewer := <-h.Register:
			h.Subscribers[viewer] = true
			log.Println("viewer", viewer.Websocket.RemoteAddr().String(), "registered at hub", h.Address.String())

		case viewer := <-h.Unregister:
			delete(h.Subscribers, viewer)
			log.Println("viewer", viewer.Websocket.RemoteAddr().String(), "unregistered from hub", h.Address.String())

			// if no subscriber left
			if len(h.Subscribers) == 0 {
				// stop poller
				log.Println("stopping poller")
				h.Poller.Quit <- true

				// remove hub
				delete(hubs, h.Address.String())

				log.Println("terminated hub")

				// end goroutine
				return
			}

		case message := <-h.Poller.Updates:
			// concurrently send message to all subscribers with a 5 second timeout
			for viewer := range h.Subscribers {
				go func(viewer *Viewer, message string) {
					select {
					case viewer.OutboundMessages <- message:
					case <-time.After(5 * time.Second):
						log.Println("forcing unregister")
						h.Unregister <- viewer
					}
				}(viewer, message)
			}
		}
	}
}
