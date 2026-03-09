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
	id string

	// hub *Hub
	hubs *HubManager

	connection *websocket.Conn
	//outbound messages buffered channel
	send chan []byte
}

type MessagePayload struct {
	Id        string `json:"id"`
	User_ID   string `json:"user_id"`
	Room_ID   string `json:"room_id"`
	Content   string `json:"content"`
	TimeStamp string `json:"timeStamp"`
	Action    string `json:"action"`
}

type ResponseMessagePayload struct {
	OriginalId string `json:"original_id"`
	Id         string `json:"id"`
	User_ID    string `json:"user_id"`
	Room_ID    string `json:"room_id"`
	Content    string `json:"content"`
	TimeStamp  string `json:"timeStamp"`
	Action     string `json:"action"`
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
		var msg MessagePayload
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error when parsing message: ", err)
			continue
		}

		hub := c.hubs.getHub(msg.Room_ID)
		if hub != nil {
			switch msg.Action {
			case "JOIN":
				//register client to that hub if not found
				if _, ok := hub.Clients[c]; !ok {
					hub.register <- c
				}
			case "SEND":
				//generate new Id
				messageId := uuid.New().String()

				responsePayload := &ResponseMessagePayload{
					OriginalId: msg.Id,
					Id:         messageId,
					User_ID:    msg.User_ID,
					Room_ID:    msg.Room_ID,
					Content:    msg.Content,
					Action:     msg.Action,
					TimeStamp:  msg.TimeStamp,
				}

				jsonPayload, err := json.Marshal(responsePayload)
				if err != nil {
					log.Println("Error when marshalling payload: ", err)
					return
				}
				hub.broadcaster <- jsonPayload
			}

		} else {
			log.Printf("Hub not found: %s", msg.Room_ID)
			log.Printf("Creating hub with id: %s", msg.Room_ID)
			c.hubs.createNewHub(msg.Room_ID)
			hub := c.hubs.getHub(msg.Room_ID)
			if hub != nil {
				hub.register <- c
			}
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

			log.Println("sending: ", string(message))

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

func serveWs(hubManager *HubManager, w http.ResponseWriter, r *http.Request, client_id string) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{id: client_id, hubs: hubManager, connection: conn, send: make(chan []byte, 256)}

	go client.handleIncomingMessages()
	go client.handleOutgoingMessages()

	log.Printf("New client connected (UUID): %s", client.id)
}
