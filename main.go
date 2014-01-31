package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"regexp"
	"text/template"
)

var (
	hubs    = map[server]hub{}
	pollers = map[server]*poller{}

	ipMatcher       = regexp.MustCompile("\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}")
	hostnameMatcher = regexp.MustCompile("localhost|([\\a-z\\.-]+)\\.([a-z\\.]{2,6})")
)

func home(resp http.ResponseWriter, req *http.Request) {
	template.Must(template.ParseFiles("html/index.html")).Execute(resp, req.Host)
}

func status(resp http.ResponseWriter, req *http.Request) {
	template.Must(template.ParseFiles("html/status.html")).Execute(resp, hubs)
}

func demo(resp http.ResponseWriter, req *http.Request) {
	template.Must(template.ParseFiles("html/embed_demo.html")).Execute(resp, req.Host)
}

func embedJS(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Println(err)
		http.NotFound(resp, req)
		return
	}

	id := req.FormValue("id")
	addr := req.FormValue("addr")
	port := req.FormValue("port")

	template.Must(template.ParseFiles("js/embed.js")).Execute(resp, struct {
		Host string
		Addr string
		Port string
		Id   string
	}{
		req.Host,
		addr,
		port,
		id,
	})
}

func main() {
	http.Handle("/style.css", http.FileServer(http.Dir("css")))
	http.Handle("/style_full.css", http.FileServer(http.Dir("css")))

	http.HandleFunc("/", home)
	http.HandleFunc("/status", status)
	http.HandleFunc("/demo", demo)
	http.HandleFunc("/embed.js", embedJS)

	http.Handle("/ws", websocket.Handler(handler))

	if err := http.ListenAndServe(":1234", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
