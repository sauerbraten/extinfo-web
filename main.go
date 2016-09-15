package main

import (
	"log"
	"net/http"
	"sync"
	"text/template"
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
	template.Must(template.ParseFiles("html/index.html")).Execute(resp, nil)
}

func detailed(resp http.ResponseWriter, req *http.Request) {
	template.Must(template.ParseFiles("html/detailed.html")).Execute(resp, nil)
}

func status(resp http.ResponseWriter, req *http.Request) {
	template.Must(template.ParseFiles("html/status.html")).Execute(resp, hubs)
}

func demo(resp http.ResponseWriter, req *http.Request) {
	template.Must(template.ParseFiles("html/embed_demo.html")).Execute(resp, PublicWebInterfaceAddress)
}

func embedJS(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println(err)
		http.NotFound(resp, req)
		return
	}

	template.Must(template.ParseFiles("js/embed.js")).Execute(resp, struct {
		Host string
		Addr string
		Port string
		Id   string
	}{
		PublicWebInterfaceAddress,
		req.FormValue("addr"),
		req.FormValue("port"),
		req.FormValue("id"),
	})
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
	r.HandleFunc("/detailed", detailed)
	r.HandleFunc("/status", status)
	r.HandleFunc("/embedding-demo", demo)
	r.HandleFunc("/embed.js", embedJS)

	r.Handle("/{fn:[-_a-z]+\\.css}", http.FileServer(http.Dir("css")))
	r.Handle("/{fn:[-_\\.a-z]+\\.js}", http.FileServer(http.Dir("js")))
	r.Handle("/{fn:[-_\\.a-z]+\\.html}", http.FileServer(http.Dir("html")))

	r.HandleFunc("/ws", websocketHandler)

	log.Println("server listening on http://localhost:8080/")
	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
