package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

// A connection from a viewer
type Viewer struct {
	*websocket.Conn
	ServerAddress Topic
	Messages      Subscriber
}

// registers websockets
func watchServer(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	addr := params.ByName("addr")

	log.Println(addr)

	addressParts := strings.Split(addr, ":")
	if len(addressParts) != 2 {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	hostname := getCanonicalHostname(addressParts[0])
	port, err := strconv.Atoi(addressParts[1])
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	topic := Topic(hostname + ":" + strconv.Itoa(port))

	err = pubsub.CreateTopicIfNotExists(topic, func() (Publisher, error) {
		return NewPollerAsPublisher(hostname, port)
	})

	if err != nil {
		log.Println(err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	messages := Subscriber(make(chan string, 1))

	pubsub.Subscribe(messages, topic)

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

	viewer.writeUpdateUntilClose()

	viewer.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(5*time.Second))
	viewer.Close()
}

// reads messages from the channel and writes them to the websocket
func (v *Viewer) writeUpdateUntilClose() {
	for message := range v.Messages {
		if err := v.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("forcing unregister:", err)
			pubsub.Unsubscribe(v.Messages, v.ServerAddress)
			break
		}
	}
}
