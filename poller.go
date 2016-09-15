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

	p.sendBasicInfoUpdates(newBasicInfo)
	p.BasicInfo = newBasicInfo

	newTeamScoresInfo, err := p.Server.GetTeamScores()
	if err != nil {
		log.Println("error getting info about team scores from server:", err)
		return err
	}

	p.sendTeamInfoUpdates(newTeamScoresInfo.Scores)
	p.TeamScores = newTeamScoresInfo.Scores

	newClientsInfo, err := p.Server.GetAllClientInfo()
	if err != nil {
		log.Println("error getting info about all clients from server:", err)
		return err
	}

	p.sendClientsInfoUpdates(newClientsInfo)
	p.ClientsInfo = newClientsInfo

	return nil
}

type ServerInfo struct {
	Description        string `json:"description"`
	MasterMode         string `json:"mastermode"`
	GameMode           string `json:"gamemode"`
	Map                string `json:"map"`
	TimeLeft           int    `json:"timeleft"`
	NumberOfClients    int    `json:"numberofclients"`
	MaxNumberOfClients int    `json:"maxnumberofclients"`
}

func (p *Poller) sendBasicInfoUpdates(newBasicInfo extinfo.BasicInfo) {
	if newBasicInfo != p.BasicInfo {
		jsonInfo, _ := json.Marshal(ServerInfo{
			Description:        newBasicInfo.Description,
			MasterMode:         newBasicInfo.MasterMode,
			GameMode:           newBasicInfo.GameMode,
			Map:                newBasicInfo.Map,
			TimeLeft:           newBasicInfo.SecsLeft,
			NumberOfClients:    newBasicInfo.NumberOfClients,
			MaxNumberOfClients: newBasicInfo.MaxNumberOfClients,
		})

		p.Updates <- "serverinfo" + "\t" + string(jsonInfo)
	}
}

type TeamInfo struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Bases []int  `json:"bases"`
}

func (p *Poller) sendTeamInfoUpdates(newTeamScores map[string]extinfo.TeamScore) {
	// handle removals
	for oldKey, oldTeamScore := range p.TeamScores {
		if _, ok := newTeamScores[oldKey]; !ok {
			p.Updates <- "delete" + "\t" + "team" + "\t" + oldTeamScore.Name
		}
	}

	// additions and updates (treated like additions)
	for _, newTeamScore := range newTeamScores {
		oldTeamScore := p.TeamScores[newTeamScore.Name]

		if newTeamScore.Name != oldTeamScore.Name || newTeamScore.Score != oldTeamScore.Score {
			jsonInfo, _ := json.Marshal(TeamInfo{
				Name:  newTeamScore.Name,
				Score: newTeamScore.Score,
			})

			p.Updates <- "team" + "\t" + string(jsonInfo)
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

func (p *Poller) sendClientsInfoUpdates(newClientInfos map[int]extinfo.ClientInfo) {
	// handle removals
	for oldKey, oldClientInfo := range p.ClientsInfo {
		if _, ok := newClientInfos[oldKey]; !ok {
			p.Updates <- "delete" + "\t" + "player" + "\t" + strconv.Itoa(oldClientInfo.ClientNum)
		}
	}

	// additions and updates (treated like additions)
	for cn, newClientInfo := range newClientInfos {
		oldClientInfo := p.ClientsInfo[cn]

		if newClientInfo.Name != oldClientInfo.Name || newClientInfo.Team != oldClientInfo.Team || newClientInfo.Frags != oldClientInfo.Frags || newClientInfo.Deaths != oldClientInfo.Deaths || newClientInfo.Accuracy != oldClientInfo.Accuracy {
			jsonInfo, _ := json.Marshal(ClientInfo{
				ClientNum: newClientInfo.ClientNum,
				Name:      newClientInfo.Name,
				Team:      newClientInfo.Team,
				Frags:     newClientInfo.Frags,
				Deaths:    newClientInfo.Deaths,
				Accuracy:  newClientInfo.Accuracy,
			})

			p.Updates <- "player" + "\t" + string(jsonInfo)
		}
	}
}
