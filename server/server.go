package main

import (
	"log"
	"sort"

	"github.com/google/uuid"
)

type Message struct {
	id      string
	content string
}

type Hub struct {
	hub_id      string
	messages    []Message
	clients     map[*Client]bool //create a map with key value pairs --- Client - bool
	broadcaster chan []byte

	register     chan *Client
	unregister   chan *Client
	unregisterId chan string
}

type HubManager struct {
	hubs map[uuid.UUID]*Hub // map value pairs --- Hub id- hub
}

func newHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[uuid.UUID]*Hub),
	}
}

func (hm *HubManager) createNewHub(hub_name string) string {
	id := uuid.New() //generate uuid for the room
	log.Println("new ID: ", id.String())
	hub := newHub(id.String())

	hm.hubs[id] = hub
	go hub.run()

	return id.String()
}

func (hm *HubManager) getHub(hub_uuid_string string) *Hub {
	//try to parse string into uuid
	hub_uuid, err := uuid.Parse(hub_uuid_string)
	if err != nil {
		log.Println("error when converting to uuid:", hub_uuid_string)
		return nil
	}
	return hm.hubs[hub_uuid]
}

func (hm *HubManager) getHubListIds() []string {
	hubsIds := make([]string, 0, len(hm.hubs))
	for id := range hm.hubs {
		hubsIds = append(hubsIds, id.String())
	}

	sort.Strings(hubsIds)

	return hubsIds
}

func newHub(id string) *Hub {
	return &Hub{
		hub_id:       id,
		messages:     make([]Message, 0),
		clients:      make(map[*Client]bool),
		broadcaster:  make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		unregisterId: make(chan string),
	}
}

func (h *Hub) disconnectClient(clientId string) {
	h.unregisterId <- clientId
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			for _, msg := range h.messages {
				select {
				case client.send <- []byte(msg.content):
				default:
					break
				}
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
		case id := <-h.unregisterId:
			for client := range h.clients {
				if client.id == id {
					close(client.send)
					delete(h.clients, client)
					break
				}
			}
		case message := <-h.broadcaster:
			for client := range h.clients {
				select {
				case client.send <- message:
					//store message
					h.messages = append(h.messages, Message{
						id:      "mock-id",
						content: string(message),
					})
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}

	}
}
