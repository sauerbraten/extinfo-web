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
	//server, err = extinfo.NewServer(addr.IP.String(), addr.Port, 5*time.Second)
	//if err != nil {
	//	return
	//}

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
	log.Println("polling started")
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
			log.Println("polling server")
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

/*
func (p *Poller) buildUpdate() (updateJSON string, err error) {
	newBasicInfo, err := p.Server.GetBasicInfo()
	if err != nil {
		err = errors.New("error getting basic info from server: " + err.Error())
		return
	}

	newTeamScoresInfo, err := p.Server.GetTeamScores()
	if err != nil {
		err = errors.New("error getting info about team scores from server:" + err.Error())
		return
	}

	newClientsInfo, err := p.Server.GetAllClientInfo()
	if err != nil {
		err = errors.New("error getting info about all clients from server:" + err.Error())
		return
	}

	var update []byte
	update, err = json.Marshal(Update{
		ServerInfo: newBasicInfo,
		Teams:      newTeamScoresInfo.Scores,
		Players:    newClientsInfo,
	})

	if err != nil {
		err = errors.New("error marshaling update:" + err.Error())
	} else {
		updateJSON = string(update)
	}

	return
}
*/

func (p *Poller) buildUpdate() (updateJSON string, err error) {
	timestamp := time.Now().Unix()

	basicInfo := extinfo.BasicInfo{
		BasicInfoRaw: extinfo.BasicInfoRaw{
			Description:        "test 123",
			Map:                "testmap",
			NumberOfClients:    5,
			MaxNumberOfClients: 7,
			SecsLeft:           int(timestamp % 600),
		},
		GameMode: "ünsta",
	}
	teamsInfo := map[string]extinfo.TeamScore{
		"good": extinfo.TeamScore{
			Name:  "good",
			Score: int(timestamp % 10),
		},
		"evil": extinfo.TeamScore{
			Name:  "evil",
			Score: int((timestamp + 4) % 10),
		},
	}
	clientsInfo := map[int]extinfo.ClientInfo{
		0: extinfo.ClientInfo{
			ClientInfoRaw: extinfo.ClientInfoRaw{
				ClientNum: 0,
				Name:      "peter",
				Team:      "good",
				Frags:     12,
				Deaths:    9,
				Accuracy:  35,
			},
		},
		1: extinfo.ClientInfo{
			ClientInfoRaw: extinfo.ClientInfoRaw{
				ClientNum: 1,
				Name:      "jane",
				Team:      "good",
				Frags:     27,
				Deaths:    32,
				Accuracy:  32,
			},
		},
		2: extinfo.ClientInfo{
			ClientInfoRaw: extinfo.ClientInfoRaw{
				ClientNum: 2,
				Name:      "hans",
				Team:      "evil",
				Frags:     34,
				Deaths:    23,
				Accuracy:  23,
			},
		},
	}

	var update []byte
	update, err = json.Marshal(Update{
		ServerInfo: basicInfo,
		Teams:      teamsInfo,
		Players:    clientsInfo,
	})

	if err != nil {
		err = errors.New("error marshaling update:" + err.Error())
	} else {
		updateJSON = string(update)
	}

	return
}
