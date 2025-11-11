package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	name string
	id   string

	// hub *Hub
	hubs *HubManager

	connection *websocket.Conn
	//outbound messages buffered channel
	send chan []byte
}

type IncomingMessage struct {
	HubId   string `json:"hubId"`
	Content string `json:"content"`
}

type MessagePayload struct {
	Username string `json:"username"`
	Content  string `json:"content"`
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
		// c.hub.broadcaster <- []byte("a client has left")
		// c.hub.unregister <- c
		c.hubs.unregisterFromAll(c)
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

		//Get message destination by parsing json
		var msg IncomingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error when parsing message: ", err)
			continue
		}

		hub := c.hubs.getHub(msg.HubId)
		if hub != nil {
			//register client to that hub if not found
			if _, ok := hub.Clients[c]; !ok {
				hub.register <- c
			}

			payload := MessagePayload{
				Username: c.name,
				Content:  msg.Content,
			}

			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				log.Println("Error when marshalling payload: ", err)
				return
			}
			hub.broadcaster <- jsonPayload
		} else {
			log.Printf("Hub not found: %s", msg.HubId)
		}
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

func serveWs(hub *HubManager, w http.ResponseWriter, r *http.Request, client_name string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	// var join_message = []byte("a new client has joined")

	if err != nil {
		log.Println(err)
		return
	}

	// create a new client
	client_uuid := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(client_name))
	client := &Client{name: client_name, id: client_uuid.String(), hubs: hub, connection: conn, send: make(chan []byte, 256)}
	// client.hub.register <- client
	// client.hub.broadcaster <- join_message

	go client.handleIncomingMessages()
	go client.handleOutgoingMessages()

	log.Printf("New client connected (UUID): %s", client.id)
}
