package main

import (
	"log"
	"net"
	"strconv"
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
			log.Println("viewer", viewer.Websocket.RemoteAddr(), "registered at hub", h.Address.String())

		case viewer := <-h.Unregister:
			delete(h.Subscribers, viewer)
			log.Println("viewer", viewer.Websocket.RemoteAddr(), "unregistered from hub", h.Address.String())

			// if no subscriber left
			if len(h.Subscribers) == 0 {
				// stop poller
				log.Println("no subscribers left, stopping poller & hub")
				h.Poller.Quit <- struct{}{}

				// remove hub
				delete(hubs, h.Address.String())

				log.Println("stopped poller & hub")

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

type Hubs map[string]*Hub

func (hubs *Hubs) GetOrCreateHub(serverAddress string) (*Hub, error) {
	// get address
	addr, err := net.ResolveUDPAddr("udp4", serverAddress)

	if err != nil {
		return nil, err
	}

	// get hostname
	hostname := addr.IP.String()

	names, err := net.LookupAddr(addr.IP.String())
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		if hub, ok := (*hubs)[name[:len(name)-1]+strconv.Itoa(addr.Port)]; ok {
			return hub, nil
		}
	}

	hostname = names[0]
	// cut off trailing '.'
	hostname = hostname[:len(hostname)-1]

	// spawn new poller and hub for new sauer server
	hub, err := newHubWithPoller(addr, hostname)
	if err != nil {
		return nil, err
	}

	(*hubs)[addr.String()] = hub

	go hub.run()

	log.Println("spawned new hub for", addr.String())

	return hub, nil
}
