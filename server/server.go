package main

import (
	"encoding/json"
	"log"
	"sort"

	"github.com/google/uuid"
)

type Hub struct {
	hub_id      string
	messages    []MessagePayload
	Clients     map[*Client]bool //create a map with key value pairs --- Client - bool
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

func (h *Hub) getClientbyId(clientId string) *Client {
	for client := range h.Clients {
		if client.id == clientId {
			return client
		}
	}

	return nil
}

func (hm *HubManager) unregisterFromAll(client *Client) {
	for _, hub := range hm.hubs {
		select {
		case hub.unregister <- client:
		default:
		}
	}
}

func (hm *HubManager) createNewHub(hub_id string) string {
	hub := newHub(hub_id)

	hub_uuid, err := uuid.Parse(hub_id)
	if err != nil {
		log.Println("error when converting to uuid:", hub_id)
		return ""
	}
	hm.hubs[hub_uuid] = hub
	go hub.run()

	return hub_id
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
		messages:     make([]MessagePayload, 0),
		Clients:      make(map[*Client]bool),
		broadcaster:  make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		unregisterId: make(chan string),
	}
}

func (h *Hub) disconnectClient(clientId string) {
	h.unregisterId <- clientId
}

func (h *Hub) isClientExists(clientId string) bool {
	// log.Println("Checking for lcient")
	for c := range h.Clients {
		if c.id == clientId {
			return true
		}
	}

	return false
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.Clients[client] = true

			for _, msg := range h.messages {
				select {
				case client.send <- []byte(msg.Content):
				default:
					break
				}
			}
		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
			}
		case id := <-h.unregisterId:
			for client := range h.Clients {
				if client.id == id {
					close(client.send)
					delete(h.Clients, client)
					break
				}
			}
		case message := <-h.broadcaster:
			var payload MessagePayload
			if err := json.Unmarshal(message, &payload); err != nil {
				log.Println("Error when extracting message payload: ", err)
				continue
			}

			for client := range h.Clients {
				select {
				case client.send <- message:
					//store message
					h.messages = append(h.messages, payload)
				default:
					close(client.send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
