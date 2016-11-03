package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"

	"encoding/json"

	"github.com/sauerbraten/extinfo"
)

type MasterServer struct {
	Publisher

	Connection    net.Conn
	ServerStates  map[string]extinfo.BasicInfo
	ServerUpdates chan Update
}

func NewMasterServerAsPublisher(addr string, notify chan<- string) (<-chan []byte, chan<- struct{}, error) {
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return nil, nil, err
	}

	publisher, updates, stop := NewPublisher(addr, notify)

	ms := &MasterServer{
		Connection: conn,
		Publisher:  publisher,
	}

	ms.refreshServers()

	go ms.loop()

	return updates, stop, nil
}

func (ms *MasterServer) loop() {
	masterErrorCount := 0
	errorCount := 0

	updateTicker := time.NewTicker(5 * time.Second)
	defer updateTicker.Stop()

	refreshTicker := time.NewTicker(30 * time.Second)
	defer refreshTicker.Stop()

	defer ms.Close()

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

func (ms *MasterServer) storeServerUpdate(upd Update) {
	serverUpdate := ServerStateUpdate{}
	err := json.Unmarshal(upd.Content, &serverUpdate)
	if err != nil {
		log.Println("error unmarshaling server state update from "+upd.Topic+":", err)
		return
	}

	ms.ServerStates[upd.Topic] = serverUpdate.ServerInfo
}

func (ms *MasterServer) refreshServers() error {
	in := bufio.NewScanner(ms.Connection)
	out := bufio.NewWriter(ms.Connection)

	// request new list

	_, err := out.WriteString("list\n")
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

		hostAndPort, err := GetCanonicalHostAndPort(msg, " ")
		if err != nil {
			log.Println("error processing addserver line ("+msg+")", err)
			continue
		}

		addr := hostAndPort.String()

		// keep old state if possible
		updatedList[addr] = ms.ServerStates[addr]

		// subscribe to updates from that server
		err = pubsub.CreateTopicIfNotExists(addr, NewPoller)
		if err != nil {
			log.Println("error creating poller for "+addr+":", err)
			continue
		}

		err = pubsub.Subscribe(ms.ServerUpdates, addr)
		if err != nil {
			log.Println("error subscribing to updates on "+addr+":", err)
			continue
		}
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
