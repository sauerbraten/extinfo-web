package main

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"text/template"
)

var hubs = map[string]*Hub{}

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

	template.Must(template.ParseFiles("js/embed.js")).Execute(resp, struct {
		Host string
		Addr string
		Port string
		Id   string
	}{
		req.Host,
		req.FormValue("addr"),
		req.FormValue("port"),
		req.FormValue("id"),
	})
}

func main() {
	http.Handle("/style.css", http.FileServer(http.Dir("css")))
	http.Handle("/style_full.css", http.FileServer(http.Dir("css")))

	http.HandleFunc("/", home)
	http.Handle("/extinfo.js", http.FileServer(http.Dir("js")))
	http.HandleFunc("/status", status)
	http.HandleFunc("/embedding-demo", demo)
	http.HandleFunc("/embed.js", embedJS)

	http.Handle("/ws", websocket.Handler(websocketHandler))

	log.Println("server listening on 0.0.0.:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
