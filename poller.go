package main

import (
	"github.com/sauerbraten/extinfo"
	"log"
	"net"
	"strconv"
	"time"
)

type Poller struct {
	Quit    chan bool
	Updates chan string
	OldInfo extinfo.BasicInfo
	Server  *extinfo.Server
}

func newPoller(addr *net.UDPAddr) (*Poller, error) {
	log.Println(addr)
	server := extinfo.NewServer(addr)
	info, err := server.GetBasicInfo()

	return &Poller{
		Quit:    make(chan bool),
		Updates: make(chan string),
		OldInfo: info,
		Server:  server,
	}, err
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
	newInfo, err := p.Server.GetBasicInfo()
	if err != nil {
		return err
	}

	p.sendBasicInfoUpdates(newInfo)

	p.OldInfo = newInfo
	return nil
}

func (p *Poller) getAllOnce() {
	p.Updates <- "timeleft\t" + strconv.Itoa(p.OldInfo.SecsLeft)
	p.Updates <- "numberofclients\t" + strconv.Itoa(p.OldInfo.NumberOfClients)
	p.Updates <- "maxnumberofclients\t" + strconv.Itoa(p.OldInfo.MaxNumberOfClients)
	p.Updates <- "map\t" + p.OldInfo.Map
	p.Updates <- "mastermode\t" + p.OldInfo.MasterMode
	p.Updates <- "gamemode\t" + p.OldInfo.GameMode
	p.Updates <- "description\t" + p.OldInfo.Description
}

func (p *Poller) sendBasicInfoUpdates(newInfo extinfo.BasicInfo) {
	// send new time
	p.Updates <- "timeleft\t" + strconv.Itoa(newInfo.SecsLeft)

	// compare other fields for changes
	if newInfo.NumberOfClients != p.OldInfo.NumberOfClients {
		p.Updates <- "numberofclients\t" + strconv.Itoa(newInfo.NumberOfClients)
	}

	if newInfo.MaxNumberOfClients != p.OldInfo.MaxNumberOfClients {
		p.Updates <- "maxnumberofclients\t" + strconv.Itoa(newInfo.MaxNumberOfClients)
	}

	if newInfo.Map != p.OldInfo.Map {
		p.Updates <- "map\t" + newInfo.Map
	}

	if newInfo.MasterMode != p.OldInfo.MasterMode {
		p.Updates <- "mastermode\t" + newInfo.MasterMode
	}

	if newInfo.GameMode != p.OldInfo.GameMode {
		p.Updates <- "gamemode\t" + newInfo.GameMode
	}

	if newInfo.Description != p.OldInfo.Description {
		p.Updates <- "description\t" + newInfo.Description
	}
}
