package main

import (
	"github.com/sauerbraten/extinfo"
	"strconv"
	"time"
)

type poller struct {
	quit     chan bool
	updates  chan string
	oldInfo  extinfo.BasicInfo
	server   *extinfo.Server
	errCount int
}

func newPoller(addr string, port int, upd chan string, qui chan bool) (*poller, error) {
	s := extinfo.NewServer(addr, port)
	info, err := s.GetBasicInfo()

	return &poller{qui, upd, info, s, 0}, err
}

func (p *poller) pollForever() {
	for {
		if p.errCount > 10 {
			return
		}

		select {
		case q := <-p.quit:
			if q {
				return
			}
		case <-time.After(5 * time.Second):
			err := p.poll()
			if err != nil {
				p.errCount++
			} else {
				p.errCount = 0
			}
		}
	}
}

func (p *poller) poll() error {
	newInfo, err := p.server.GetBasicInfo()
	if err != nil {
		return err
	}

	p.sendBasicInfoUpdates(newInfo)

	p.oldInfo = newInfo
	return nil
}

func (p *poller) getOnce() {
	p.updates <- "timeleft\t" + strconv.Itoa(p.oldInfo.SecsLeft)
	p.updates <- "numberofclients\t" + strconv.Itoa(p.oldInfo.NumberOfClients)
	p.updates <- "maxnumberofclients\t" + strconv.Itoa(p.oldInfo.MaxNumberOfClients)
	p.updates <- "map\t" + p.oldInfo.Map
	p.updates <- "mastermode\t" + p.oldInfo.MasterMode
	p.updates <- "gamemode\t" + p.oldInfo.GameMode
	p.updates <- "description\t" + p.oldInfo.Description
}

func (p *poller) sendBasicInfoUpdates(newInfo extinfo.BasicInfo) {
	// send new time
	p.updates <- "timeleft\t" + strconv.Itoa(newInfo.SecsLeft)

	// compare other fields for changes
	if newInfo.NumberOfClients != p.oldInfo.NumberOfClients {
		p.updates <- "numberofclients\t" + strconv.Itoa(newInfo.NumberOfClients)
	}

	if newInfo.MaxNumberOfClients != p.oldInfo.MaxNumberOfClients {
		p.updates <- "maxnumberofclients\t" + strconv.Itoa(newInfo.MaxNumberOfClients)
	}

	if newInfo.Map != p.oldInfo.Map {
		p.updates <- "map\t" + newInfo.Map
	}

	if newInfo.MasterMode != p.oldInfo.MasterMode {
		p.updates <- "mastermode\t" + newInfo.MasterMode
	}

	if newInfo.GameMode != p.oldInfo.GameMode {
		p.updates <- "gamemode\t" + newInfo.GameMode
	}

	if newInfo.Description != p.oldInfo.Description {
		p.updates <- "description\t" + newInfo.Description
	}
}
