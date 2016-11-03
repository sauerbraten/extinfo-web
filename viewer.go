package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

// A connection from a viewer
type Viewer struct {
	*websocket.Conn
	ServerAddress string
	Messages      chan Update
}

// reads messages from the channel and writes them to the websocket
func (v *Viewer) writeUpdatesUntilClose() {
	for message := range v.Messages {
		if err := v.WriteMessage(websocket.TextMessage, message.Content); err != nil {
			log.Println("sending failed: forcing unregister for viewer", v.Messages)
			return
		}
	}
}

// handles websocket connections subscribing for server state updates
func watchServer(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	addr := params.ByName("addr")
	topic := addr + " (detailed)"

	log.Println(req.RemoteAddr, "started watching", addr)

	subscribeWebsocket(resp, req, topic, func(publisher Publisher) error {
		log.Println("starting to poll", addr)
		return NewPoller(
			publisher,
			func(p *Poller) { p.WithPlayers = true },
			func(p *Poller) { p.WithTeams = true },
			func(p *Poller) { p.Address = addr },
		)
	})

	log.Println(req.RemoteAddr, "stopped watching", addr)
}

func watchMaster(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Println(req.RemoteAddr, "started watching the master server list")

	subscribeWebsocket(resp, req, DefaultMasterServerAddress, func(publisher Publisher) error {
		log.Println("starting to poll the master server")
		NewMasterServerAsPublisher(publisher, func(ms *MasterServer) { ms.ServerAddress = DefaultMasterServerAddress })
		return nil
	})

	log.Println(req.RemoteAddr, "stopped watching the master server list")
}

func subscribeWebsocket(resp http.ResponseWriter, req *http.Request, topic string, useNewPublisher func(Publisher) error) {
	err := pubsub.CreateTopicIfNotExists(topic, useNewPublisher)
	if err != nil {
		log.Println("creating poller for "+topic+" failed:", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	messages := make(chan Update)

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	viewer := &Viewer{
		Conn:          conn,
		ServerAddress: topic,
		Messages:      messages,
	}

	pubsub.Subscribe(messages, topic)

	viewer.writeUpdatesUntilClose()

	pubsub.Unsubscribe(messages, topic)

	_ = viewer.Close()
}
