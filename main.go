package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

const PublicWebInterfaceAddress = "extinfo.sauerworld.org"

var (
	pubsub *PubSub = &PubSub{
		Publishers:    map[Topic]Publisher{},
		Subscriptions: map[Topic]map[Subscriber]bool{},
	}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func init() {
	go pubsub.Loop()
}

func main() {
	r := httprouter.New()
	r.RedirectTrailingSlash = true

	r.GET("/", home)

	r.ServeFiles("/css/*filepath", http.Dir("css"))
	r.ServeFiles("/js/*filepath", http.Dir("js"))

	//r.GET("/master", watchMasterServerList)
	r.GET("/server/:addr", watchServer)

	log.Println("server listening on http://localhost:8080/")
	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func home(resp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.ServeFile(resp, req, "html/index.html")
}

func getCanonicalHostname(hostname string) string {
	names, err := net.LookupAddr(hostname)
	if err != nil {
		return hostname
	}

	return names[0][:len(hostname)-1] // cut off trailing '.'
}
