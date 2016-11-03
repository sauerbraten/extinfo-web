package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Publisher

	Server        *extinfo.Server
	Configuration <-chan func(*Poller)
	WithTeams     bool
	WithPlayers   bool
}

func NewConfigurablePollerAsPublisher(addr string, notify chan<- string) (<-chan []byte, chan<- struct{}, chan<- func(*Poller), error) {
	hostAndPort, err := HostAndPortFromString(addr, ":")
	if err != nil {
		return nil, nil, nil, err
	}

	server, err := extinfo.NewServer(hostAndPort.Host, hostAndPort.Port, 5*time.Second)
	if err != nil {
		return nil, nil, nil, err
	}

	updates := make(chan []byte, 1)
	stop := make(chan struct{})
	conf := make(chan func(*Poller))

	poller := &Poller{
		Server:        server,
		Configuration: conf,
		Publisher: Publisher{
			Topic:        addr,
			NotifyPubSub: notify,
			Updates:      updates,
			Stop:         stop,
		},
	}

	err = poller.poll()
	if err != nil {
		return nil, nil, nil, err
	}

	go poller.pollForever()

	return updates, stop, conf, nil
}

func NewPollerAsPublisher(addr string, notify chan<- string) (<-chan []byte, chan<- struct{}, error) {
	upd, stop, _, err := NewConfigurablePollerAsPublisher(addr, notify)
	return upd, stop, err
}

func (p *Poller) pollForever() {
	errorCount := 0
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer close(p.Updates)

	for {
		select {
		case <-ticker.C:
			err := p.poll()
			if err != nil {
				log.Println(err)
				errorCount++
				if errorCount > 10 {
					log.Println("problem with server, stopping poller")
					return
				}
			} else {
				errorCount = 0
			}
		case <-p.Stop:
			return
		case configuration := <-p.Configuration:
			configuration(p)
		}
	}
}

func (p *Poller) poll() error {
	update, err := p.buildUpdate()
	if err != nil {
		return err
	}

	p.NotifyPubSub <- p.Topic
	p.Updates <- update

	return nil
}

type ServerStateUpdate struct {
	ServerInfo extinfo.BasicInfo            `json:"serverinfo"`
	Teams      map[string]extinfo.TeamScore `json:"teams"`
	Players    map[int]extinfo.ClientInfo   `json:"players"`
}

func (p *Poller) buildUpdate() ([]byte, error) {
	update := ServerStateUpdate{}
	var err error

	update.ServerInfo, err = p.Server.GetBasicInfo()
	if err != nil {
		return nil, errors.New("error getting basic info from server: " + err.Error())
	}

	if p.WithTeams {
		teams, err := p.Server.GetTeamScores()
		if err != nil {
			return nil, errors.New("error getting info about team scores from server: " + err.Error())
		}
		update.Teams = teams.Scores
	}

	if p.WithPlayers {
		update.Players, err = p.Server.GetAllClientInfo()
		if err != nil {
			return nil, errors.New("error getting info about all clients from server: " + err.Error())
		}
	}

	return json.Marshal(update)
}
