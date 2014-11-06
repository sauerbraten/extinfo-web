package main

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Quit        chan struct{}
	Updates     chan string
	BasicInfo   extinfo.BasicInfo
	TeamScores  map[string]extinfo.TeamScore
	ClientsInfo map[int]extinfo.ClientInfo
	Server      *extinfo.Server
}

func newPoller(addr *net.UDPAddr) (p *Poller, err error) {
	var server *extinfo.Server
	server, err = extinfo.NewServer(addr.IP.String(), addr.Port, 5*time.Second)
	if err != nil {
		return
	}

	var basicInfo extinfo.BasicInfo
	basicInfo, err = server.GetBasicInfo()
	if err != nil {
		return
	}

	var scoresInfo extinfo.TeamScores
	scoresInfo, err = server.GetTeamScores()
	if err != nil {
		return
	}

	var clientsInfo map[int]extinfo.ClientInfo
	clientsInfo, err = server.GetAllClientInfo()
	if err != nil {
		return
	}

	p = &Poller{
		Quit:        make(chan struct{}),
		Updates:     make(chan string),
		BasicInfo:   basicInfo,
		TeamScores:  scoresInfo.Scores,
		ClientsInfo: clientsInfo,
		Server:      server,
	}

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
	newBasicInfo, err := p.Server.GetBasicInfo()
	if err != nil {
		log.Println("error getting basic info from server:", err)
		return err
	}

	newTeamScoresInfo, err := p.Server.GetTeamScores()
	if err != nil {
		log.Println("error getting info about team scores from server:", err)
		return err
	}

	newClientsInfo, err := p.Server.GetAllClientInfo()
	if err != nil {
		log.Println("error getting info about all clients from server:", err)
		return err
	}

	p.sendBasicInfoUpdates(newBasicInfo, true)
	p.BasicInfo = newBasicInfo

	p.sendTeamScoresUpdates(newTeamScoresInfo.Scores, true)
	p.TeamScores = newTeamScoresInfo.Scores

	p.sendClientsInfoUpdates(newClientsInfo, true)
	p.ClientsInfo = newClientsInfo

	return nil
}

func (p *Poller) sendBasicInfoUpdates(newBasicInfo extinfo.BasicInfo, onlyNew bool) {
	if !onlyNew || newBasicInfo.Description != p.BasicInfo.Description {
		p.Updates <- "description\t" + newBasicInfo.Description
	}

	if !onlyNew || newBasicInfo.GameMode != p.BasicInfo.GameMode {
		p.Updates <- "gamemode\t" + newBasicInfo.GameMode
	}

	if !onlyNew || newBasicInfo.Map != p.BasicInfo.Map {
		p.Updates <- "map\t" + newBasicInfo.Map
	}

	if !onlyNew || newBasicInfo.NumberOfClients != p.BasicInfo.NumberOfClients {
		p.Updates <- "numberofclients\t" + strconv.Itoa(newBasicInfo.NumberOfClients)
	}

	if !onlyNew || newBasicInfo.MaxNumberOfClients != p.BasicInfo.MaxNumberOfClients {
		p.Updates <- "maxnumberofclients\t" + strconv.Itoa(newBasicInfo.MaxNumberOfClients)
	}

	if !onlyNew || newBasicInfo.MasterMode != p.BasicInfo.MasterMode {
		p.Updates <- "mastermode\t" + newBasicInfo.MasterMode
	}

	if !onlyNew || newBasicInfo.SecsLeft != p.BasicInfo.SecsLeft {
		p.Updates <- "timeleft\t" + strconv.Itoa(newBasicInfo.SecsLeft)
	}
}

func (p *Poller) sendTeamScoresUpdates(newTeamScores map[string]extinfo.TeamScore, onlyNew bool) {
	for _, newTeamScore := range newTeamScores {
		oldTeamScore, ok := p.TeamScores[newTeamScore.Name]
		if !onlyNew || !ok || newTeamScore.Score != oldTeamScore.Score {
			p.Updates <- "team\t" + newTeamScore.Name + "\t" + strconv.Itoa(newTeamScore.Score)
		}
	}
}

type ClientInfo struct {
	ClientNum int    `json:"cn"`
	Name      string `json:"name"`
	Team      string `json:"team"`
	Frags     int    `json:"frags"`
	Deaths    int    `json:"deaths"`
	Accuracy  int    `json:"accuracy"`
}

func (p *Poller) sendClientsInfoUpdates(newClientsInfo map[int]extinfo.ClientInfo, onlyNew bool) {
	clientInfos := []ClientInfo{}

	for _, clientInfo := range newClientsInfo {
		clientInfos = append(clientInfos, ClientInfo{
			clientInfo.ClientNum,
			clientInfo.Name,
			clientInfo.Team,
			clientInfo.Frags,
			clientInfo.Deaths,
			clientInfo.Accuracy,
		})
	}

	clientInfosJSON, err := json.Marshal(clientInfos)
	if err != nil {
		println(err)
	}

	p.Updates <- "players\t" + string(clientInfosJSON)

}
