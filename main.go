package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

const PublicWebInterfaceAddress = "extinfo.sauerworld.org"

var (
	pubsub *PubSub

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func init() {
	pubsub = NewPubSub()
	go pubsub.Loop()
}

func main() {
	r := httprouter.New()
	r.RedirectTrailingSlash = true

	r.GET("/", home)

	r.ServeFiles("/css/*filepath", http.Dir("css"))
	r.ServeFiles("/js/*filepath", http.Dir("js"))

	r.GET("/master", watchMaster)
	r.GET("/server/:addr", watchServer)

	log.Println("server listening on http://localhost:8080/")
	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func home(resp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.ServeFile(resp, req, "html/index.html")
}
