package main

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/sauerbraten/extinfo"
)

type Poller struct {
	Quit      chan struct{}
	Updates   chan string
	BasicInfo extinfo.BasicInfo
	Server    *extinfo.Server
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
		Quit:      make(chan struct{}),
		Updates:   make(chan string),
		BasicInfo: info,
		Server:    server,
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
	return nil
}

func (p *Poller) getAllOnce() {
	p.Updates <- "description\t" + p.BasicInfo.Description
	p.Updates <- "gamemode\t" + p.BasicInfo.GameMode
	p.Updates <- "map\t" + p.BasicInfo.Map
	p.Updates <- "numberofclients\t" + strconv.Itoa(p.BasicInfo.NumberOfClients)
	p.Updates <- "maxnumberofclients\t" + strconv.Itoa(p.BasicInfo.MaxNumberOfClients)
	p.Updates <- "mastermode\t" + p.BasicInfo.MasterMode
	p.Updates <- "timeleft\t" + strconv.Itoa(p.BasicInfo.SecsLeft)
}

func (p *Poller) sendBasicInfoUpdates(newBasicInfo extinfo.BasicInfo) {
	if newBasicInfo.Description != p.BasicInfo.Description {
		p.Updates <- "description\t" + newBasicInfo.Description
	}

	if newBasicInfo.GameMode != p.BasicInfo.GameMode {
		p.Updates <- "gamemode\t" + newBasicInfo.GameMode
	}

	if newBasicInfo.Map != p.BasicInfo.Map {
		p.Updates <- "map\t" + newBasicInfo.Map
	}

	if newBasicInfo.NumberOfClients != p.BasicInfo.NumberOfClients {
		p.Updates <- "numberofclients\t" + strconv.Itoa(newBasicInfo.NumberOfClients)
	}

	if newBasicInfo.MaxNumberOfClients != p.BasicInfo.MaxNumberOfClients {
		p.Updates <- "maxnumberofclients\t" + strconv.Itoa(newBasicInfo.MaxNumberOfClients)
	}

	if newBasicInfo.MasterMode != p.BasicInfo.MasterMode {
		p.Updates <- "mastermode\t" + newBasicInfo.MasterMode
	}

	if newBasicInfo.SecsLeft != p.BasicInfo.SecsLeft {
		p.Updates <- "timeleft\t" + strconv.Itoa(newBasicInfo.SecsLeft)
	}
}
