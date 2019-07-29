package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

func main() {
	addr := flag.String("addr", "", "Address to bind server on")
	flag.Parse()

	room := newRoom()

	http.Handle("/", &rootHandler{fname: "chat.html"})
	http.Handle("/room", room)

	go room.run()

	log.Printf("Listening on %v\n", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type rootHandler struct {
	once  sync.Once
	fname string
	templ *template.Template
}

func (t *rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.fname)))
	})
	t.templ.Execute(w, r)
}
