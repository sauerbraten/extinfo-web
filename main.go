package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const PublicWebInterfaceAddress = "extinfo.sauerworld.org"

var hubs = map[string]*Hub{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func home(resp http.ResponseWriter, req *http.Request) {
	http.ServeFile(resp, req, "html/index.html")
}

// registers websockets
func websocketHandler(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	v := &Viewer{
		Websocket:        conn,
		OutboundMessages: make(chan string),
	}

	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func(wg *sync.WaitGroup) {
		v.writeUntilClose()
		wg.Done()
	}(wg)

	go func(wg *sync.WaitGroup) {
		v.readUntilClose()
		wg.Done()
	}(wg)

	wg.Wait()

	v.Websocket.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(5*time.Second))
	v.Websocket.Close()
}

func main() {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", home)

	r.Handle("/{fn:[-_\\.a-z]+\\.css}", http.FileServer(http.Dir("css")))
	r.Handle("/{fn:[-_\\.a-z]+\\.js}", http.FileServer(http.Dir("js")))
	r.Handle("/{fn:[-_\\.a-z]+\\.html}", http.FileServer(http.Dir("html")))

	r.HandleFunc("/ws", websocketHandler)

	log.Println("server listening on http://localhost:8080/")
	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
