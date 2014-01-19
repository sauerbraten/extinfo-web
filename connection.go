package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net"
	"strconv"
	"strings"
)

type connection struct {
	ws   *websocket.Conn // the websocket connection
	send chan string     // buffered channel of outbound messages
}

// reads from the websocket
func (c *connection) reader() {
	for {
		var message string
		err := websocket.Message.Receive(c.ws, &message)
		if err != nil {
			break
		}
		// subscribe to a hub

		// get sauer address and port
		parts := strings.Split(message, "\t")
		addr := strings.ToLower(parts[0])
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Println(err)
		}

		ip := ""
		host := ""

		// get domain or IP, whichever is missing
		if hostnameMatcher.MatchString(addr) {
			// hostname given, get IP
			IPs, err := net.LookupHost(addr)
			if err != nil {
				log.Println(err)
				return
			}

			ip = IPs[0]
			host = addr
		} else if ipMatcher.MatchString(addr) {
			// IP given, get hostname
			names, err := net.LookupAddr(addr)
			if err != nil {
				log.Println(err)
			}

			ip = addr
			host = names[0]
			// cut off '.' at the end
			host = host[:len(host)-1]
		} else {
			// invalid addr argument, return
			return
		}

		s := server{ip, host, port}
		var p *poller

		h, ok := hubs[s]
		if !ok {
			// spawn new poller and hub for new sauer server
			upd := make(chan string)
			qui := make(chan bool, 1)

			p, err = newPoller(addr, port, upd, qui)
			if err != nil {
				log.Println(err)
				return
			}

			pollers[s] = p
			hubs[s] = hub{map[*connection]bool{}, upd, qui, make(chan *connection), make(chan *connection), s}

			go hubs[s].run()
			go p.pollForever()

			h = hubs[s]

			log.Println("spawned new poller for", addr, "port", port)
		} else {
			p = pollers[s]
		}

		log.Println("websocket registered at hub", s.IP, "port", port)

		h.register <- c
		defer func() { h.unregister <- c }()
		p.getOnce()
	}

	c.ws.Close()
}

// reads from the channel and writes to the websocket
func (c *connection) writer() {
	for message := range c.send {
		err := websocket.Message.Send(c.ws, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

// registers websockets
func handler(ws *websocket.Conn) {
	c := &connection{send: make(chan string, 256), ws: ws}

	go c.writer()
	c.reader()
}
