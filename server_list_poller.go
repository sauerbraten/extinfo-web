package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"

	"encoding/json"

	"github.com/sauerbraten/extinfo"
	"github.com/sauerbraten/pubsub"
)

const DefaultMasterServerAddress = "sauerbraten.org:28787"

type serverListEntryUpdate struct {
	Address string
	Update  []byte
}

// ServerListPoller polls the master server and publishes updates about the server list by
// starting a poller for the basic info of each server on the list and subscribing to its updates.
type ServerListPoller struct {
	*pubsub.Publisher
	subscriptions map[string]chan []byte // server address → update channel

	MasterServerAddress string

	serverStates  map[string]extinfo.BasicInfo
	serverUpdates chan serverListEntryUpdate // all update channels are merged into this channel
}

func NewServerListPoller(publisher *pubsub.Publisher, conf ...func(*ServerListPoller)) {
	slp := &ServerListPoller{
		Publisher:     publisher,
		serverStates:  map[string]extinfo.BasicInfo{},
		serverUpdates: make(chan serverListEntryUpdate),
		subscriptions: map[string]chan []byte{},
	}

	for _, configFunc := range conf {
		configFunc(slp)
	}

	go slp.loop()
}

// refresh list and update clients, then do both periodically forever
func (slp *ServerListPoller) loop() {
	defer slp.Close()
	defer func() {
		for addr, updates := range slp.subscriptions {
			broker.Unsubscribe(updates, addr)
		}
	}()
	defer debug("stopped polling the master server list")

	debug("started polling the master server list")

	err := slp.refreshServers()
	if err != nil {
		log.Println("error getting initial master server list:", err)
		return
	}

	err = slp.publishUpdate()
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
			err := slp.refreshServers()
			if err != nil {
				log.Println(err)
				masterErrorCount++
				if masterErrorCount > 10 {
					log.Println("problem with master server, exiting server list poller loop")
					return
				}
			} else {
				masterErrorCount = 0
			}

		case <-updateTicker.C:
			err := slp.publishUpdate()
			if err != nil {
				log.Println(err)
				errorCount++
				if errorCount > 10 {
					log.Println("problem with updates, exiting server list poller loop")
					return
				}
			} else {
				errorCount = 0
			}

		case supd := <-slp.serverUpdates:
			slp.storeServerUpdate(supd)

		case <-slp.Stop:
			return
		}
	}
}

func (slp *ServerListPoller) storeServerUpdate(supd serverListEntryUpdate) {
	serverUpdate := ServerStateUpdate{}
	err := json.Unmarshal(supd.Update, &serverUpdate)
	if err != nil {
		log.Println("error unmarshaling server state update from "+supd.Address+":", err)
		return
	}

	slp.serverStates[supd.Address] = serverUpdate.ServerInfo
}

func (slp *ServerListPoller) refreshServers() error {
	// open connection
	conn, err := net.DialTimeout("tcp", slp.MasterServerAddress, 15*time.Second)
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
		oldState, known := slp.serverStates[addr]
		updatedList[addr] = oldState

		if !known {
			// subscribe to updates if the server is new
			updates, publisher := broker.Subscribe(addr)
			if publisher != nil {
				err = NewServerPoller(
					publisher,
					func(sp *ServerPoller) { sp.Address = addr },
				)
				if err != nil {
					return err
				}
			}

			slp.subscriptions[addr] = updates

			// merge updates from all servers into one channel to select on
			go func(addr string, updates <-chan []byte) {
				for upd := range updates {
					slp.serverUpdates <- serverListEntryUpdate{
						Address: addr,
						Update:  upd,
					}
				}
			}(addr, updates)
		}
	}
	if err := in.Err(); err != nil {
		return err
	}

	// unsubscribe from updates about servers not on the list anymore
	for addr := range slp.serverStates {
		if _, ok := updatedList[addr]; !ok {
			broker.Unsubscribe(slp.subscriptions[addr], addr)
			delete(slp.subscriptions, addr)
		}
	}

	slp.serverStates = updatedList

	return nil
}

type serverListEntry struct {
	Address string `json:"address"`
	extinfo.BasicInfo
}

func (slp *ServerListPoller) publishUpdate() error {
	serverList := []serverListEntry{}
	for addr, state := range slp.serverStates {
		if state.NumberOfClients <= 0 {
			continue
		}
		serverList = append(serverList, serverListEntry{
			Address:   addr,
			BasicInfo: state,
		})
	}

	update, err := json.Marshal(serverList)
	if err != nil {
		return err
	}

	slp.Publish(update)

	return nil
}
