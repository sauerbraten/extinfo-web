package master

import (
	"log"
	"time"

	"github.com/sauerbraten/pubsub"
)

type ServerListPublisher struct {
	p               *pubsub.Publisher
	s               *Server
	servers         ServerList
	refreshInterval time.Duration
}

func NewServerListPublisher(p *pubsub.Publisher, s *Server, refreshInterval time.Duration) {
	slp := &ServerListPublisher{
		p:               p,
		s:               s,
		servers:         ServerList{},
		refreshInterval: refreshInterval,
	}

	go slp.loop()
}

func (slp *ServerListPublisher) loop() {
	refreshTicker := time.NewTicker(slp.refreshInterval)
	defer refreshTicker.Stop()

	errorCount := 0

	for {
		select {
		case <-refreshTicker.C:
			newList, err := slp.s.ServerList()
			if err != nil {
				// TODO: exponential backoff
				log.Println(err)
				errorCount++
				if errorCount > 100 {
					log.Println("problem with master server, exiting server list polling loop")
					return
				}
			} else {
				slp.publishChanges(newList)
				errorCount = 0
			}
		case <-slp.p.Stop:
			return
		}
	}
}

func (slp *ServerListPublisher) publishChanges(newList ServerList) {
	for addr := range newList {
		if _, ok := slp.servers[addr]; !ok {
			slp.p.Publish([]byte("add " + addr))
		}
	}

	for addr := range slp.servers {
		if _, ok := newList[addr]; !ok {
			slp.p.Publish([]byte("del " + addr))
		}
	}

	slp.servers = newList
}

func (slp *ServerListPublisher) republishAll() {
	slp.p.Publish([]byte("clear"))

	for addr := range slp.servers {
		slp.p.Publish([]byte("add " + addr))
	}
}
