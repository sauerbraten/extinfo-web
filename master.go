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

type MasterServer struct {
	pubsub.Publisher

	Address       string
	ServerStates  map[string]extinfo.BasicInfo
	ServerUpdates chan pubsub.Update
}

func NewMasterServerAsPublisher(publisher pubsub.Publisher, conf ...func(*MasterServer)) {
	ms := &MasterServer{
		Publisher:     publisher,
		ServerStates:  map[string]extinfo.BasicInfo{},
		ServerUpdates: make(chan pubsub.Update),
	}

	for _, configFunc := range conf {
		configFunc(ms)
	}

	go ms.loop()
}

// refresh list and update clients, then do both periodically forever
func (ms *MasterServer) loop() {
	ms.refreshServers()
	ms.update()

	masterErrorCount := 0
	errorCount := 0

	updateTicker := time.NewTicker(5 * time.Second)
	defer updateTicker.Stop()

	refreshTicker := time.NewTicker(30 * time.Second)
	defer refreshTicker.Stop()

	defer ms.Close()
	defer func() {
		for topic := range ms.ServerStates {
			pubsub.Unsubscribe(ms.ServerUpdates, topic)
		}
	}()

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

func (ms *MasterServer) storeServerUpdate(upd pubsub.Update) {
	serverUpdate := ServerStateUpdate{}
	err := json.Unmarshal(upd.Content, &serverUpdate)
	if err != nil {
		log.Println("error unmarshaling server state update from "+upd.Topic+":", err)
		return
	}

	ms.ServerStates[upd.Topic] = serverUpdate.ServerInfo
}

func (ms *MasterServer) refreshServers() error {
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

	// receive new list
	updatedList := map[string]extinfo.BasicInfo{}

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
		updatedList[addr] = ms.ServerStates[addr]

		// subscribe to updates from that server
		updates, err := pubsub.Subscribe(addr, func(publisher pubsub.Publisher) error {
			return NewPoller(
				publisher,
				func(p *Poller) { p.Address = addr },
			)
		})
		if err != nil {
			log.Println("error subscribing to updates on "+addr+":", err)
			continue
		}

		// proxy updates from all servers into single channel
		go func(updates <-chan pubsub.Update) {
			for update := range updates {
				ms.ServerUpdates <- update
			}
		}(updates)
	}

	ms.ServerStates = updatedList

	return in.Err()
}

func (ms *MasterServer) update() error {
	update, err := ms.buildUpdate()
	if err != nil {
		return err
	}

	ms.Publish(update)

	return nil
}

type ServerListEntry struct {
	Address string `json:"address"`
	extinfo.BasicInfo
}

func (ms *MasterServer) buildUpdate() ([]byte, error) {
	serverList := []ServerListEntry{}
	for addr, state := range ms.ServerStates {
		serverList = append(serverList, ServerListEntry{
			Address:   addr,
			BasicInfo: state,
		})
	}

	return json.Marshal(serverList)
}
