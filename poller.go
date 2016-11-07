package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sauerbraten/extinfo"
	"github.com/sauerbraten/extinfo-web/internal/pubsub"
)

type ServerPoller struct {
	pubsub.Publisher

	Server      *extinfo.Server
	Address     string
	WithTeams   bool
	WithPlayers bool
}

func NewServerPoller(publisher pubsub.Publisher, config ...func(*ServerPoller)) error {
	sp := &ServerPoller{
		Publisher: publisher,
	}

	for _, configFunc := range config {
		configFunc(sp)
	}

	host, port, err := HostAndPortFromString(sp.Address, ":")
	if err != nil {
		return err
	}

	sp.Server, err = extinfo.NewServer(host, port, 5*time.Second)
	if err != nil {
		return err
	}

	go sp.loop()

	return nil
}

// poll once immediately, then periodically
func (sp *ServerPoller) loop() {
	defer sp.Close()

	err := sp.update()
	if err != nil {
		log.Println("initial poll failed:", err)
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
				log.Println(err)
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
	ServerInfo extinfo.BasicInfo            `json:"serverinfo"`
	Teams      map[string]extinfo.TeamScore `json:"teams,omitempty"`
	Players    map[int]extinfo.ClientInfo   `json:"players,omitempty"`
}

func (sp *ServerPoller) update() error {
	update := ServerStateUpdate{}
	var err error

	update.ServerInfo, err = sp.Server.GetBasicInfo()
	if err != nil {
		return errors.New("error getting basic info from server: " + err.Error())
	}

	if sp.WithTeams {
		teams, err := sp.Server.GetTeamScores()
		if err != nil {
			return errors.New("error getting info about team scores from server: " + err.Error())
		}
		update.Teams = teams.Scores
	}

	if sp.WithPlayers {
		update.Players, err = sp.Server.GetAllClientInfo()
		if err != nil {
			return errors.New("error getting info about all clients from server: " + err.Error())
		}
	}

	updateJSON, err := json.Marshal(update)
	if err != nil {
		return err
	}

	sp.Publish(updateJSON)

	return nil
}
