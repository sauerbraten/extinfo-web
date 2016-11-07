package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"

	"encoding/json"

	"github.com/sauerbraten/extinfo"
	"github.com/sauerbraten/extinfo-web/internal/pubsub"
)

const DefaultMasterServerAddress = "sauerbraten.org:28787"

type MasterServerPoller struct {
	pubsub.Publisher

	Address       string
	ServerStates  map[string]extinfo.BasicInfo
	ServerUpdates chan pubsub.Update
	Subscriptions map[string]chan pubsub.Update
}

func NewMasterServerPoller(publisher pubsub.Publisher, conf ...func(*MasterServerPoller)) {
	ms := &MasterServerPoller{
		Publisher:     publisher,
		ServerStates:  map[string]extinfo.BasicInfo{},
		ServerUpdates: make(chan pubsub.Update),
		Subscriptions: map[string]chan pubsub.Update{},
	}

	for _, configFunc := range conf {
		configFunc(ms)
	}

	go ms.loop()
}

// refresh list and update clients, then do both periodically forever
func (ms *MasterServerPoller) loop() {
	defer ms.Close()
	defer func() {
		for addr, updates := range ms.Subscriptions {
			pubsub.Unsubscribe(updates, addr)
		}
	}()

	err := ms.refreshServers()
	if err != nil {
		log.Println("error getting initial master server list:", err)
		return
	}

	err = ms.update()
	if err != nil {
		log.Println("error publishing first update:", err)
		return
	}

	masterErrorCount := 0
	errorCount := 0

	updateTicker := time.NewTicker(5 * time.Second)
	defer updateTicker.Stop()

	refreshTicker := time.NewTicker(30 * time.Second)
	defer refreshTicker.Stop()

	for {
		select {
		case <-refreshTicker.C:
			err := ms.refreshServers()
			if err != nil {
				log.Println(err)
				masterErrorCount++
				if masterErrorCount > 10 {
					log.Println("problem with master server, exiting master server loop")
					return
				}
			} else {
				masterErrorCount = 0
			}

		case <-updateTicker.C:
			err := ms.update()
			if err != nil {
				log.Println(err)
				errorCount++
				if errorCount > 10 {
					log.Println("problem with updates, exiting master server loop")
					return
				}
			} else {
				errorCount = 0
			}

		case upd := <-ms.ServerUpdates:
			ms.storeServerUpdate(upd)

		case <-ms.Stop:
			return
		}
	}
}

func (ms *MasterServerPoller) storeServerUpdate(upd pubsub.Update) {
	serverUpdate := ServerStateUpdate{}
	err := json.Unmarshal(upd.Content, &serverUpdate)
	if err != nil {
		log.Println("error unmarshaling server state update from "+upd.Topic+":", err)
		return
	}

	ms.ServerStates[upd.Topic] = serverUpdate.ServerInfo
}

func (ms *MasterServerPoller) refreshServers() error {
	// open connection
	conn, err := net.DialTimeout("tcp", ms.Address, 15*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	in := bufio.NewScanner(conn)
	out := bufio.NewWriter(conn)

	// request new list
	_, err = out.WriteString("list\n")
	if err != nil {
		return err
	}

	err = out.Flush()
	if err != nil {
		return err
	}

	updatedList := map[string]extinfo.BasicInfo{}

	// process response
	for in.Scan() {
		msg := in.Text()
		if msg == "\x00" {
			// end of list
			continue
		}

		msg = strings.TrimPrefix(msg, "addserver ")
		msg = strings.TrimSpace(msg)

		addr := strings.Replace(msg, " ", ":", -1)

		// keep old state if possible
		oldState, known := ms.ServerStates[addr]
		updatedList[addr] = oldState

		if !known {
			// subscribe to updates if the server is new
			updates, err := pubsub.Subscribe(addr, func(publisher pubsub.Publisher) error {
				return NewServerPoller(
					publisher,
					func(sp *ServerPoller) { sp.Address = addr },
				)
			})
			if err != nil {
				log.Println("error subscribing to updates on "+addr+":", err)
			}

			ms.Subscriptions[addr] = updates

			go func(updates <-chan pubsub.Update) {
				for upd := range updates {
					ms.ServerUpdates <- upd
				}
			}(updates)
		}
	}

	// unsubscribe from updates about servers not on the list anymore
	for addr := range ms.ServerStates {
		if _, ok := updatedList[addr]; !ok {
			pubsub.Unsubscribe(ms.ServerUpdates, addr)
			delete(ms.Subscriptions, addr)
		}
	}

	ms.ServerStates = updatedList

	return in.Err()
}

type ServerListEntry struct {
	Address string `json:"address"`
	extinfo.BasicInfo
}

func (ms *MasterServerPoller) update() error {
	serverList := []ServerListEntry{}
	for addr, state := range ms.ServerStates {
		serverList = append(serverList, ServerListEntry{
			Address:   addr,
			BasicInfo: state,
		})
	}

	update, err := json.Marshal(serverList)
	if err != nil {
		return err
	}

	ms.Publish(update)

	return nil
}
