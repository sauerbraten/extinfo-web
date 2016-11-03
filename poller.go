package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Server         *extinfo.Server
	Updates        chan<- string
	LastupdateJSON string
	Stop           <-chan struct{}
}

func NewPollerAsPublisher(hostname string, port int) (<-chan string, chan<- struct{}, error) {
	server, err := extinfo.NewServer(hostname, port, 5*time.Second)
	if err != nil {
		return nil, nil, err
	}

	updates := make(chan string, 1)
	stop := make(chan struct{})
	poller := &Poller{
		Server:  server,
		Updates: updates,
		Stop:    stop,
	}

	poller.poll()
	go poller.pollForever()

	return updates, stop, nil
}

func (p *Poller) pollForever() {
	errorCount := 0
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			if errorCount > 10 {
				log.Println("problem with server, stopping poller")
				close(p.Updates)
				ticker.Stop()
				return
			}

			err := p.poll()

			if err != nil {
				log.Println(err)
				errorCount++
			} else {
				errorCount = 0
			}
		case <-p.Stop:
			close(p.Updates)
			ticker.Stop()
			return
		}
	}
}

func (p *Poller) poll() (err error) {
	p.LastupdateJSON, err = p.buildUpdate()
	if err != nil {
		return
	}

	p.Updates <- p.LastupdateJSON
	return
}

type Update struct {
	ServerInfo extinfo.BasicInfo            `json:"serverinfo"`
	Teams      map[string]extinfo.TeamScore `json:"teams"`
	Players    map[int]extinfo.ClientInfo   `json:"players"`
}

func (p *Poller) buildUpdate() (updateJSON string, err error) {
	basicInfo, err := p.Server.GetBasicInfo()
	if err != nil {
		err = errors.New("error getting basic info from server: " + err.Error())
		return
	}

	teamScoresInfo, err := p.Server.GetTeamScores()
	if err != nil {
		err = errors.New("error getting info about team scores from server:" + err.Error())
		return
	}

	clientsInfo, err := p.Server.GetAllClientInfo()
	if err != nil {
		err = errors.New("error getting info about all clients from server:" + err.Error())
		return
	}

	var update []byte
	update, err = json.Marshal(Update{
		ServerInfo: basicInfo,
		Teams:      teamScoresInfo.Scores,
		Players:    clientsInfo,
	})

	if err != nil {
		err = errors.New("error marshaling update:" + err.Error())
	} else {
		updateJSON = string(update)
	}

	return
}
