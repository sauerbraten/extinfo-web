package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sauerbraten/chef/pkg/extinfo"
	"github.com/sauerbraten/pubsub"
)

type ServerPoller struct {
	*pubsub.Publisher[[]byte]

	server      *extinfo.Server
	host        string
	port        int
	WithTeams   bool
	WithPlayers bool
}

func NewServerPoller(publisher *pubsub.Publisher[[]byte], config ...func(*ServerPoller)) error {
	sp := &ServerPoller{
		Publisher: publisher,
	}

	for _, configFunc := range config {
		configFunc(sp)
	}

	_server, err := extinfo.NewServer(pinger, sp.host, sp.port, 10*time.Second)
	if err != nil {
		return err
	}
	sp.server = _server

	go sp.loop()

	return nil
}

// poll once immediately, then periodically
func (sp *ServerPoller) loop() {
	defer sp.Close()
	defer debug("stopped polling", sp.Topic())

	debug("started polling", sp.Topic())

	err := sp.update()
	if err != nil {
		log.Println("fetch initial server state:", err)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	errorCount := 0

	for {
		select {
		case <-ticker.C:
			err := sp.update()
			if err != nil {
				log.Println("fetch server state:", err)
				errorCount++
				if errorCount > 10 {
					log.Println("problem with server, exiting loop")
					return
				}
			} else {
				errorCount = 0
			}

		case <-sp.Stop:
			return
		}
	}
}

type ServerStateUpdate struct {
	ServerInfo *extinfo.BasicInfo           `json:"serverinfo"`
	Teams      map[string]extinfo.TeamScore `json:"teams,omitempty"`
	Players    map[int]*extinfo.ClientInfo  `json:"players,omitempty"`
	Mod        extinfo.ServerMod            `json:"mod,omitempty"`
}

func (sp *ServerPoller) update() error {
	update := ServerStateUpdate{}
	var err error

	update.ServerInfo, err = sp.server.GetBasicInfo()
	if err != nil {
		return errors.New("fetch basic info from server: " + err.Error())
	}

	update.Mod, err = sp.server.GetServerMod()
	if err != nil {
		log.Printf("detecting mod of %s:%d: %v", sp.host, sp.port, err.Error())
	}

	if sp.WithTeams && update.ServerInfo.GameMode.IsTeamMode() {
		teams, err := sp.server.GetTeamScores()
		if err != nil {
			return errors.New("fetch info about team scores from server: " + err.Error())
		}
		update.Teams = teams.Scores
	}

	if sp.WithPlayers {
		update.Players, err = sp.server.GetClientInfo(-1)
		if err != nil {
			return errors.New("fetch info about all clients from server: " + err.Error())
		}
	}

	updateJSON, err := json.Marshal(update)
	if err != nil {
		return err
	}

	sp.Publish(updateJSON)

	return nil
}
