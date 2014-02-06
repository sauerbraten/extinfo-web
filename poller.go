package main

import (
	"github.com/sauerbraten/extinfo"
	"log"
	"net"
	"strconv"
	"time"
)

type Poller struct {
	Quit            chan bool
	BasicUpdates    chan string
	ExtendedUpdates chan string
	oldBasicInfo    extinfo.BasicInfo
	oldPlayerInfos  map[int]extinfo.PlayerInfo
	Server          *extinfo.Server
}

func newPoller(addr *net.UDPAddr) (*Poller, error) {
	server := extinfo.NewServer(addr)

	basicInfo, err := server.GetBasicInfo()
	if err != nil {
		return nil, err
	}

	playerInfos, err := server.GetAllPlayerInfo()
	if err != nil {
		return nil, err
	}

	return &Poller{
		Quit:            make(chan bool),
		BasicUpdates:    make(chan string),
		ExtendedUpdates: make(chan string),
		oldBasicInfo:    basicInfo,
		oldPlayerInfos:  playerInfos,
		Server:          server,
	}, nil
}

func (p *Poller) pollForever() {
	t := time.NewTicker(5 * time.Second)
	errorCount := 0
	for {
		if errorCount > 10 {
			t.Stop()
			<-p.Quit
			return
		}

		select {
		case <-p.Quit:
			t.Stop()
			return

		case <-t.C:
			err := p.pollForBasicInfo()
			if err != nil {
				log.Println(err)
				errorCount++
			} else {
				errorCount = 0
			}

			err = p.pollForPlayerInfos()
			if err != nil {
				log.Println(err)
				errorCount++
			} else {
				errorCount = 0
			}
		}
	}
}

func (p *Poller) pollForBasicInfo() error {
	newBasicInfo, err := p.Server.GetBasicInfo()
	if err != nil {
		return err
	}

	p.sendBasicUpdates(newBasicInfo)

	p.oldBasicInfo = newBasicInfo
	return nil
}

func (p *Poller) pollForPlayerInfos() error {
	newPlayerInfos, err := p.Server.GetAllPlayerInfo()
	if err != nil {
		return err
	}

	p.sendExtendedUpdates(newPlayerInfos)

	p.oldPlayerInfos = newPlayerInfos
	return nil
}

func (p *Poller) getAllInfoOnce() {
	p.BasicUpdates <- "timeleft\t" + strconv.Itoa(p.oldBasicInfo.SecsLeft)
	p.BasicUpdates <- "numberofclients\t" + strconv.Itoa(p.oldBasicInfo.NumberOfClients)
	p.BasicUpdates <- "maxnumberofclients\t" + strconv.Itoa(p.oldBasicInfo.MaxNumberOfClients)
	p.BasicUpdates <- "map\t" + p.oldBasicInfo.Map
	p.BasicUpdates <- "mastermode\t" + p.oldBasicInfo.MasterMode
	p.BasicUpdates <- "gamemode\t" + p.oldBasicInfo.GameMode
	p.BasicUpdates <- "description\t" + p.oldBasicInfo.Description

	for _, playerInfo := range p.oldPlayerInfos {
		p.sendCompletePlayerStats(playerInfo)
	}
}

func (p *Poller) sendBasicUpdates(newBasicInfo extinfo.BasicInfo) {
	// send new time
	p.BasicUpdates <- "timeleft\t" + strconv.Itoa(newBasicInfo.SecsLeft)

	// compare other fields for changes
	if newBasicInfo.NumberOfClients != p.oldBasicInfo.NumberOfClients {
		p.BasicUpdates <- "numberofclients\t" + strconv.Itoa(newBasicInfo.NumberOfClients)
	}

	if newBasicInfo.MaxNumberOfClients != p.oldBasicInfo.MaxNumberOfClients {
		p.BasicUpdates <- "maxnumberofclients\t" + strconv.Itoa(newBasicInfo.MaxNumberOfClients)
	}

	if newBasicInfo.Map != p.oldBasicInfo.Map {
		p.BasicUpdates <- "map\t" + newBasicInfo.Map
	}

	if newBasicInfo.MasterMode != p.oldBasicInfo.MasterMode {
		p.BasicUpdates <- "mastermode\t" + newBasicInfo.MasterMode
	}

	if newBasicInfo.GameMode != p.oldBasicInfo.GameMode {
		p.BasicUpdates <- "gamemode\t" + newBasicInfo.GameMode
	}

	if newBasicInfo.Description != p.oldBasicInfo.Description {
		p.BasicUpdates <- "description\t" + newBasicInfo.Description
	}
}

func (p *Poller) sendExtendedUpdates(newExtendedInfo interface{}) {
	switch info := newExtendedInfo.(type) {
	case map[int]extinfo.PlayerInfo:
		for _, player := range info {
			p.sendPlayerStatsUpdates(player)
		}

		// check for disconnected clients
		for cn, _ := range p.oldPlayerInfos {
			if _, ok := info[cn]; !ok {
				p.ExtendedUpdates <- "playerstats\t" + strconv.Itoa(cn) + "\tdisconnected\t1"
			}
		}
	}
}

func (p *Poller) sendPlayerStatsUpdates(newPlayerInfo extinfo.PlayerInfo) {
	cn := newPlayerInfo.ClientNum
	old, ok := p.oldPlayerInfos[cn]
	if !ok {
		p.sendCompletePlayerStats(newPlayerInfo)
		return
	}
	prefix := "playerstats\t" + strconv.Itoa(cn) + "\t"

	if newPlayerInfo.State != old.State {
		p.ExtendedUpdates <- prefix + "state\t" + newPlayerInfo.State
	}

	if newPlayerInfo.Team != old.Team {
		p.ExtendedUpdates <- prefix + "team\t" + newPlayerInfo.Team
	}

	if newPlayerInfo.Name != old.Name {
		p.ExtendedUpdates <- prefix + "name\t" + newPlayerInfo.Name
	}

	if newPlayerInfo.Frags != old.Frags {
		p.ExtendedUpdates <- prefix + "frags\t" + strconv.Itoa(newPlayerInfo.Frags)
	}

	if newPlayerInfo.Deaths != old.Deaths {
		p.ExtendedUpdates <- prefix + "deaths\t" + strconv.Itoa(newPlayerInfo.Deaths)
	}
}

func (p *Poller) sendCompletePlayerStats(playerInfo extinfo.PlayerInfo) {
	prefix := "playerstats\t" + strconv.Itoa(playerInfo.ClientNum) + "\t"

	p.ExtendedUpdates <- prefix + "state\t" + playerInfo.State
	p.ExtendedUpdates <- prefix + "team\t" + playerInfo.Team
	p.ExtendedUpdates <- prefix + "name\t" + playerInfo.Name
	p.ExtendedUpdates <- prefix + "frags\t" + strconv.Itoa(playerInfo.Frags)
	p.ExtendedUpdates <- prefix + "deaths\t" + strconv.Itoa(playerInfo.Deaths)
}
