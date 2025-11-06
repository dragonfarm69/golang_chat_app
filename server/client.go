package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // Time allowed to write message
	pongWait       = 60 * time.Second    // Time allowed to read messages
	pingPeriod     = (pongWait * 9) / 10 // must be less than pongWait - ping sent period
	maxMessageSize = 51515
)

var (
	NEW_LINE = []byte("\n")
)

type Client struct {
	// name string
	id string

	hub *Hub

	connection *websocket.Conn

	//outbound messages buffered channel
	send chan []byte
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *Client) handleIncomingMessages() {
	defer func() {
		c.hub.broadcaster <- []byte("a client has left")
		c.hub.unregister <- c
		c.connection.Close()
	}()
	c.connection.SetReadLimit(maxMessageSize)
	c.connection.SetReadDeadline(time.Now().Add(pongWait))
	c.connection.SetPongHandler(func(string) error { c.connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.connection.ReadMessage()
		if err != nil {
			log.Print("error ReadMessage: ", err)
			break
		}
		log.Println("message:", string(message))

		//broadcast message to all users in the same hub
		c.hub.broadcaster <- message
	}
}

func (c *Client) handleOutgoingMessages() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// log.Println("receiving: ", string(message))

			w.Write(message)
			w.Write(NEW_LINE)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
				w.Write(NEW_LINE)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	var join_message = []byte("a new client has joined")

	if err != nil {
		log.Println(err)
		return
	}

	// create a new client
	client := &Client{id: "test", hub: hub, connection: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	client.hub.broadcaster <- join_message

	go client.handleIncomingMessages()
	go client.handleOutgoingMessages()
}
