package main

import (
	"io"
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

// reads from the websocket connection until an error occurs, then returns
// (necessary for the websocket package to process the 'close' control frame sent by the client.)
func (v *Viewer) readFramesUntilError() {
	for {
		_, _, err := v.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				debug(err)
			}
			return
		}
	}
}

// reads messages from the channel and writes them to the websocket
func (v *Viewer) writeUpdatesUntilClose() {
	for update := range v.Updates {
		if err := v.WriteMessage(websocket.TextMessage, update); err != nil {
			debug("sending failed: forcing unsubscribe for viewer", v.RemoteAddr())
			return
		}
	}
}

// handles websocket connections subscribing for server state updates
func watchServer(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	addr := params.ByName("addr")
	topic := addr + " (detailed)"

	host, port, err := hostAndPort(addr)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		io.WriteString(resp, err.Error())
		return
	}

	log.Println(req.RemoteAddr, "started watching", topic)

	watch(resp, req, topic, func(publisher *pubsub.Publisher) error {
		return NewServerPoller(
			publisher,
			func(sp *ServerPoller) { sp.WithPlayers = true },
			func(sp *ServerPoller) { sp.WithTeams = true },
			func(sp *ServerPoller) { sp.host = host; sp.port = port },
		)
	})

	log.Println(req.RemoteAddr, "stopped watching", topic)
}

func watchMaster(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	log.Println(req.RemoteAddr, "started watching the master server list")

	watch(resp, req, DefaultMasterServerAddress, func(publisher *pubsub.Publisher) error {
		NewServerListPoller(publisher)
		return nil
	})

	log.Println(req.RemoteAddr, "stopped watching the master server list")
}

func watch(resp http.ResponseWriter, req *http.Request, topic string, useNewPublisher func(*pubsub.Publisher) error) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	updates, newPublisher := broker.Subscribe(topic)
	if newPublisher != nil {
		err := useNewPublisher(newPublisher)
		if err != nil {
			log.Println(err)
			return
		}
	}

	viewer := &Viewer{
		Conn:          conn,
		ServerAddress: topic,
		Updates:       updates,
	}
	defer viewer.Close()

	go viewer.readFramesUntilError()
	viewer.writeUpdatesUntilClose()

	broker.Unsubscribe(updates, topic)
}
