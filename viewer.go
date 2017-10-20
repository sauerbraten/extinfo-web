package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/sauerbraten/pubsub"
)

// A connection from a viewer
type Viewer struct {
	*websocket.Conn
	ServerAddress string
	Updates       <-chan []byte
}

// reads messages from the channel and writes them to the websocket
func (v *Viewer) writeUpdatesUntilClose() {
	for update := range v.Updates {
		if err := v.WriteMessage(websocket.TextMessage, update); err != nil {
			log.Println("sending failed: forcing unregister for viewer", v.RemoteAddr())
			return
		}
	}
}

// handles websocket connections subscribing for server state updates
func watchServer(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	addr := params.ByName("addr")
	topic := addr + " (detailed)"

	log.Println(req.RemoteAddr, "started watching", topic)

	subscribeWebsocket(resp, req, topic, func(publisher *pubsub.Publisher) error {
		log.Println("starting to poll", addr)
		return NewServerPoller(
			publisher,
			func(sp *ServerPoller) { sp.WithPlayers = true },
			func(sp *ServerPoller) { sp.WithTeams = true },
			func(sp *ServerPoller) { sp.Address = addr },
		)
	})

	log.Println(req.RemoteAddr, "stopped watching", topic)
}

func watchMaster(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Println(req.RemoteAddr, "started watching the master server list")

	subscribeWebsocket(resp, req, DefaultMasterServerAddress, func(publisher *pubsub.Publisher) error {
		log.Println("starting to poll the master server")
		NewServerListPoller(publisher, func(msp *ServerListPoller) { msp.MasterServerAddress = DefaultMasterServerAddress })
		return nil
	})

	log.Println(req.RemoteAddr, "stopped watching the master server list")
}

func subscribeWebsocket(resp http.ResponseWriter, req *http.Request, topic string, useNewPublisher func(*pubsub.Publisher) error) {
	updates, newPublisher := broker.Subscribe(topic)
	if newPublisher != nil {
		err := useNewPublisher(newPublisher)
		if err != nil {
			log.Println(err)
			return
		}
	}

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	viewer := &Viewer{
		Conn:          conn,
		ServerAddress: topic,
		Updates:       updates,
	}

	viewer.writeUpdatesUntilClose()

	broker.Unsubscribe(updates, topic)

	_ = viewer.Close()
}
