package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"encoding/json"

	"github.com/sauerbraten/chef/pkg/extinfo"
	"github.com/sauerbraten/chef/pkg/master"
	"github.com/sauerbraten/pubsub"
)

const DefaultMasterServerAddress = "sauerbraten.org:28787"

type serverListEntryUpdate struct {
	Address string
	Update  []byte
}

type serverState struct {
	*extinfo.BasicInfo
	Mod string `json:"mod"`
}

// ServerListPoller polls the master server and publishes updates about the server list by
// starting a poller for the basic info of each server on the list and subscribing to its updates.
type ServerListPoller struct {
	*pubsub.Publisher
	subscriptions map[string]chan []byte // server address → update channel

	ms *master.Server

	serverStates  map[string]serverState
	serverUpdates chan serverListEntryUpdate // all update channels are merged into this channel
}

func NewServerListPoller(publisher *pubsub.Publisher) {
	slp := &ServerListPoller{
		Publisher:     publisher,
		subscriptions: map[string]chan []byte{},

		ms: master.New(DefaultMasterServerAddress, 15*time.Second),

		serverStates:  map[string]serverState{},
		serverUpdates: make(chan serverListEntryUpdate),
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

	slp.serverStates[supd.Address] = serverState{
		serverUpdate.ServerInfo,
		serverUpdate.Mod,
	}
}

func (slp *ServerListPoller) refreshServers() error {
	servers, err := slp.ms.ServerList()
	if err != nil {
		return err
	}

	updatedList := map[string]serverState{}

	// process response
	for _, addr := range servers {
		// keep old state if possible
		oldState, known := slp.serverStates[addr]
		updatedList[addr] = oldState

		if !known {
			// subscribe to updates if the server is new
			updates, publisher := broker.Subscribe(addr)
			if publisher != nil {
				host, port, err := hostAndPort(addr)
				if err != nil {
					return err
				}
				err = NewServerPoller(
					publisher,
					func(sp *ServerPoller) { sp.host = host; sp.port = port },
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
	serverState
}

func (slp *ServerListPoller) publishUpdate() error {
	serverList := []serverListEntry{}
	for addr, state := range slp.serverStates {
		if state.BasicInfo == nil || state.NumberOfClients <= 0 {
			continue
		}
		serverList = append(serverList, serverListEntry{
			Address:     addr,
			serverState: state,
		})
	}

	update, err := json.Marshal(serverList)
	if err != nil {
		return err
	}

	slp.Publish(update)

	return nil
}

func hostAndPort(addr string) (string, int, error) {
	host, _port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", -1, fmt.Errorf("parsing '%s' as host:port tuple: %v", addr, err)
	}

	port, err := strconv.Atoi(_port)
	if err != nil {
		return "", -1, fmt.Errorf("error converting port '%s' to int: %v", _port, err)
	}

	return host, port, nil
}
