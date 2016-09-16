package main

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Quit           chan struct{}
	Updates        chan string
	LastupdateJSON string
	Server         *extinfo.Server
}

type Update struct {
	ServerInfo extinfo.BasicInfo            `json:"serverinfo"`
	Teams      map[string]extinfo.TeamScore `json:"teams"`
	Players    map[int]extinfo.ClientInfo   `json:"players"`
}

func newPoller(addr *net.UDPAddr) (p *Poller, err error) {
	var server *extinfo.Server
	server, err = extinfo.NewServer(addr.IP.String(), addr.Port, 5*time.Second)
	if err != nil {
		return
	}

	p = &Poller{
		Quit:    make(chan struct{}),
		Updates: make(chan string),
		Server:  server,
	}

	var updateJSON string
	updateJSON, err = p.buildUpdate()
	if err != nil {
		return
	}

	p.LastupdateJSON = updateJSON

	return
}

func (p *Poller) pollForever() {
	t := time.NewTicker(5 * time.Second)
	errorCount := 0
	for {
		if errorCount > 10 {
			log.Println("problem with server, stopping poller")
			p.Quit <- struct{}{}
		}

		select {
		case <-p.Quit:
			t.Stop()
			return

		case <-t.C:
			err := p.poll()
			if err != nil {
				log.Println(err)
				errorCount++
			} else {
				errorCount = 0
			}
		}
	}
}

func (p *Poller) poll() error {
	updateJSON, err := p.buildUpdate()
	if err == nil {
		p.Updates <- updateJSON
	}

	return err
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
