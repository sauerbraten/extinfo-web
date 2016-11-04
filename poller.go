package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/sauerbraten/extinfo"
	"github.com/sauerbraten/extinfo-web/internal/pubsub"
)

type Poller struct {
	pubsub.Publisher

	Server      *extinfo.Server
	Address     string
	WithTeams   bool
	WithPlayers bool
}

func NewPoller(publisher pubsub.Publisher, config ...func(*Poller)) error {
	poller := &Poller{
		Publisher: publisher,
	}

	for _, configFunc := range config {
		configFunc(poller)
	}

	host, port, err := HostAndPortFromString(poller.Address, ":")
	if err != nil {
		return err
	}

	poller.Server, err = extinfo.NewServer(host, port, 5*time.Second)
	if err != nil {
		return err
	}

	go poller.loop()

	return nil
}

// poll once immediately, then periodically
func (p *Poller) loop() {
	p.poll()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer p.Close()

	for {
		select {
		case <-ticker.C:
			err := p.poll()
			if err != nil {
				// don't print errors, there are too many...
				return
			}
		case <-p.Stop:
			return
		}
	}
}

func (p *Poller) poll() error {
	update, err := p.buildUpdate()
	if err != nil {
		return err
	}

	p.Publish(update)

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
