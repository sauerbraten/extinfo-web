package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Publisher

	Server      *extinfo.Server
	Address     string
	WithTeams   bool
	WithPlayers bool
}

func NewPoller(publisher Publisher, config ...func(*Poller)) error {
	poller := &Poller{
		Publisher: publisher,
	}

	for _, configFunc := range config {
		configFunc(poller)
	}

	hostAndPort, err := HostAndPortFromString(poller.Address, ":")
	if err != nil {
		return err
	}

	poller.Server, err = extinfo.NewServer(hostAndPort.Host, hostAndPort.Port, 5*time.Second)
	if err != nil {
		return err
	}

	go poller.poll()
	go poller.pollForever()

	return nil
}

func (p *Poller) pollForever() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer p.Close()

	for {
		select {
		case <-ticker.C:
			err := p.poll()
			if err != nil {
				// don't print errors, there's too many...
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
