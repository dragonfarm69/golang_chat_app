package app

import (
	"chat-app-server/Internal/data"
	shared "chat-app-server/Shared"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleIncomingMessages(c *Client, storage *data.DataStorage) {
	defer func() {
		// c.hub.broadcaster <- []byte("a client has left")
		// c.hub.unregister <- c
		c.Hubs.UnregisterFromAll(c)
		c.Connection.Close()
	}()
	c.Connection.SetReadLimit(maxMessageSize)
	c.Connection.SetReadDeadline(time.Now().Add(pongWait))
	c.Connection.SetPongHandler(func(string) error { c.Connection.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			log.Print("error ReadMessage: ", err)
			break
		}
		log.Println("message:", string(message))

		//Get message destination by parsing json
		var msg shared.MessagePayload
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error when parsing message: ", err)
			continue
		}

		hub := c.Hubs.GetHub(msg.Room_ID)
		if hub != nil {
			switch msg.Action {
			case "JOIN":
				//register client to that hub if not found
				if _, ok := hub.Clients[c]; !ok {
					hub.Register <- c

					//save to db

				}
			case "TYPING":
				responsePayload := &shared.ResponseMessagePayload{
					OriginalId: msg.Id,
					Id:         msg.Id,
					User_ID:    msg.User_ID,
					UserName:   msg.UserName,
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
				hub.Broadcaster <- jsonPayload
			case "STOP_TYPING":
				responsePayload := &shared.ResponseMessagePayload{
					OriginalId: msg.Id,
					Id:         msg.Id,
					User_ID:    msg.User_ID,
					UserName:   msg.UserName,
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
				hub.Broadcaster <- jsonPayload
			case "SEND":
				//generate new Id
				messageId := ulid.Make().String()

				responsePayload := &shared.ResponseMessagePayload{
					OriginalId: msg.Id,
					Id:         messageId,
					User_ID:    msg.User_ID,
					UserName:   msg.UserName,
					Room_ID:    msg.Room_ID,
					Content:    msg.Content,
					Action:     msg.Action,
					TimeStamp:  msg.TimeStamp,
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				//change the id before saving to db
				msg.Id = messageId
				id, err := storage.AddNewMessageToDB(ctx, msg)
				if err != nil {
					log.Println(err)
					return
				}

				log.Println("Saved message successfully with id: ", id)

				storage.AddNewMessageToRedis(ctx, msg)

				jsonPayload, err := json.Marshal(responsePayload)
				if err != nil {
					log.Println("Error when marshalling payload: ", err)
					return
				}
				hub.Broadcaster <- jsonPayload
			}

		} else {
			log.Printf("Hub not found: %s", msg.Room_ID)
			log.Printf("Creating hub with id: %s", msg.Room_ID)
			c.Hubs.CreateNewHub(msg.Room_ID)
			hub := c.Hubs.GetHub(msg.Room_ID)
			if hub != nil {
				hub.Register <- c
			}
		}
	}
}

func handleOutgoingMessages(c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			log.Println("Sending: ", string(message))

			w.Write(message)
			w.Write(NEW_LINE)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
				w.Write(NEW_LINE)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hubManager *HubManager, w http.ResponseWriter, r *http.Request, client_id string, storage *data.DataStorage) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{Id: client_id, Hubs: hubManager, Connection: conn, Send: make(chan []byte, 256)}

	go handleIncomingMessages(client, storage)
	go handleOutgoingMessages(client)

	log.Printf("New client connected (UUID): %s", client.Id)
}
