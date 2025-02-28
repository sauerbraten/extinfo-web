package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sauerbraten/chef/pkg/extinfo"
	"github.com/sauerbraten/pubsub"
)

//go:embed html css js
var static embed.FS

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var broker = pubsub.NewBroker[[]byte]()

var pinger *extinfo.Pinger

func init() {
	var err error
	pinger, err = extinfo.NewPinger(":33333")
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("GET /", home)
	http.HandleFunc("GET /master", watchMaster)
	http.HandleFunc("GET /server/{addr}", watchServer)
	http.Handle("GET /css/", http.FileServerFS(static))
	http.Handle("GET /js/", http.FileServerFS(static))

	log.Println("server listening on http://localhost:8080/")
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func home(resp http.ResponseWriter, req *http.Request) {
	http.ServeFileFS(resp, req, static, "html/index.html")
}

func debug(a ...interface{}) {
	//log.Println(a...)
}
