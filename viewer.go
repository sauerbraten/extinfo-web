package main

import (
	"log"
	"net/http"
	"time"

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
			log.Println("forcing unregister:", err)
			pubsub.Unsubscribe(v.Messages, v.ServerAddress)
			break
		}
	}
}

// handles websocket connections subscribing for server state updates
func watchServer(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	subscribeWebsocket(resp, req, params.ByName("addr"), func(addr string, notify chan<- string) (<-chan []byte, chan<- struct{}, error) {
		updates, stop, conf, err := NewConfigurablePollerAsPublisher(addr, notify)
		if err == nil {
			conf <- func(p *Poller) { p.WithPlayers = true }
			conf <- func(p *Poller) { p.WithTeams = true }
		}
		return updates, stop, err
	})
}

func watchMaster(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	subscribeWebsocket(resp, req, "sauerbraten.org:28787", NewMasterServerAsPublisher)
}

func subscribeWebsocket(resp http.ResponseWriter, req *http.Request, topic string, createPublisher func(string, chan<- string) (<-chan []byte, chan<- struct{}, error)) {
	err := pubsub.CreateTopicIfNotExists(topic, createPublisher)
	if err != nil {
		log.Println("creating publisher for "+topic+" failed:", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer pubsub.stopPublisherIfNoSubs(topic)

	messages := make(chan Update, 1)

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

	viewer.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(5*time.Second))
	viewer.Close()
}
