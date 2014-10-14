package main

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Quit    chan bool
	Updates chan string
	OldInfo extinfo.BasicInfo
	Server  *extinfo.Server
}

func newPoller(addr *net.UDPAddr) (p *Poller, err error) {
	var server *extinfo.Server
	server, err = extinfo.NewServer(addr.IP.String(), addr.Port, 5*time.Second)
	if err != nil {
		return
	}

	var info extinfo.BasicInfo
	info, err = server.GetBasicInfo()
	if err != nil {
		return
	}

	p = &Poller{
		Quit:    make(chan bool),
		Updates: make(chan string),
		OldInfo: info,
		Server:  server,
	}

	return
}

func (p *Poller) pollForever() {
	t := time.NewTicker(5 * time.Second)
	errorCount := 0
	for {
		if errorCount > 10 {
			log.Println("problem with server, stopping poller")
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
		log.Println("error getting basic info from server:", err)
		return err
	}

	p.sendBasicInfoUpdates(newInfo)

	p.OldInfo = newInfo
	return nil
}

func (p *Poller) getAllOnce() {
	p.Updates <- "description\t" + p.OldInfo.Description
	p.Updates <- "gamemode\t" + p.OldInfo.GameMode
	p.Updates <- "map\t" + p.OldInfo.Map
	p.Updates <- "numberofclients\t" + strconv.Itoa(p.OldInfo.NumberOfClients)
	p.Updates <- "maxnumberofclients\t" + strconv.Itoa(p.OldInfo.MaxNumberOfClients)
	p.Updates <- "mastermode\t" + p.OldInfo.MasterMode
	p.Updates <- "timeleft\t" + strconv.Itoa(p.OldInfo.SecsLeft)
}

func (p *Poller) sendBasicInfoUpdates(newInfo extinfo.BasicInfo) {
	if newInfo.Description != p.OldInfo.Description {
		p.Updates <- "description\t" + newInfo.Description
	}

	if newInfo.GameMode != p.OldInfo.GameMode {
		p.Updates <- "gamemode\t" + newInfo.GameMode
	}

	if newInfo.Map != p.OldInfo.Map {
		p.Updates <- "map\t" + newInfo.Map
	}

	if newInfo.NumberOfClients != p.OldInfo.NumberOfClients {
		p.Updates <- "numberofclients\t" + strconv.Itoa(newInfo.NumberOfClients)
	}

	if newInfo.MaxNumberOfClients != p.OldInfo.MaxNumberOfClients {
		p.Updates <- "maxnumberofclients\t" + strconv.Itoa(newInfo.MaxNumberOfClients)
	}

	if newInfo.MasterMode != p.OldInfo.MasterMode {
		p.Updates <- "mastermode\t" + newInfo.MasterMode
	}

	if newInfo.SecsLeft != p.OldInfo.SecsLeft {
		p.Updates <- "timeleft\t" + strconv.Itoa(newInfo.SecsLeft)
	}
}
