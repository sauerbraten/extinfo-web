package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/sauerbraten/extinfo"
	"github.com/sauerbraten/pubsub"
)

type ServerPoller struct {
	*pubsub.Publisher

	server      *extinfo.Server
	Address     *net.UDPAddr
	WithTeams   bool
	WithPlayers bool
}

func NewServerPoller(publisher *pubsub.Publisher, config ...func(*ServerPoller)) error {
	sp := &ServerPoller{
		Publisher: publisher,
	}

	for _, configFunc := range config {
		configFunc(sp)
	}

	_server, err := extinfo.NewServer(*sp.Address, 10*time.Second)
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
	Mod        string                       `json:"mod,omitempty"`
}

func (sp *ServerPoller) update() error {
	update := ServerStateUpdate{}
	var err error

	update.ServerInfo, err = sp.server.GetBasicInfo()
	if err != nil {
		return errors.New("error getting basic info from server: " + err.Error())
	}

	update.Mod, err = sp.server.GetServerMod()
	if err != nil {
		log.Printf("error detecting mod of %s: %v", sp.Address, err.Error())
	}

	if sp.WithTeams && extinfo.IsTeamMode(update.ServerInfo.GameMode) {
		teams, err := sp.server.GetTeamScores()
		if err != nil {
			return errors.New("error getting info about team scores from server: " + err.Error())
		}
		update.Teams = teams.Scores
	}

	if sp.WithPlayers {
		update.Players, err = sp.server.GetAllClientInfo()
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

func hostAndPort(addr string) (string, int, error) {
	host, _port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", -1, errors.New("could not split address into host and port: " + err.Error())
	}

	port, err := strconv.Atoi(_port)
	if err != nil {
		return "", -1, errors.New("could not convert port to int: " + err.Error())
	}

	return host, port, nil
}
