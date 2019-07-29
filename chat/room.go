package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var (
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  socketBufferSize,
		WriteBufferSize: socketBufferSize,
	}
)

type room struct {
	broadcastCh chan []byte
	joinCh      chan *client
	leaveCh     chan *client
	clients     map[*client]bool
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sock, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	cl := &client{
		socket: sock,
		sendCh: make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.joinCh <- cl
	defer func() { r.leaveCh <- cl }()
	go cl.write()
	cl.read()
}

func (r *room) run() {
	for {
		select {
		case client := <-r.joinCh:
			r.clients[client] = true
		case client := <-r.leaveCh:
			delete(r.clients, client)
			close(client.sendCh)
		case msg := <-r.broadcastCh:
			for client := range r.clients {
				client.sendCh <- msg
			}
		}
	}
}

func newRoom() *room {
	return &room{
		broadcastCh: make(chan []byte),
		joinCh:      make(chan *client),
		leaveCh:     make(chan *client),
		clients:     map[*client]bool{},
	}
}
