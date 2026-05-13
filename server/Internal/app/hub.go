package app

import (
	shared "chat-app-server/Shared"
	"encoding/json"
	"log"
	"sort"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Id string

	// hub *Hub
	Hubs *HubManager

	Connection *websocket.Conn
	//outbound messages buffered channel
	Send chan []byte
}

type Hub struct {
	Hub_id      string
	Messages    []shared.MessagePayload
	Clients     map[*Client]bool //create a map with key value pairs --- Client - bool
	Broadcaster chan []byte

	Register     chan *Client
	Unregister   chan *Client
	UnregisterId chan string
}

type HubManager struct {
	Hubs map[uuid.UUID]*Hub // map value pairs --- Hub id- hub
}

func newHubManager() *HubManager {
	return &HubManager{
		Hubs: make(map[uuid.UUID]*Hub),
	}
}

func newHub(id string) *Hub {
	return &Hub{
		Hub_id:       id,
		Messages:     make([]shared.MessagePayload, 0),
		Clients:      make(map[*Client]bool),
		Broadcaster:  make(chan []byte),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		UnregisterId: make(chan string),
	}
}

func (h *HubManager) CreateNewHub(hub_id string) string {
	hub := newHub(hub_id)

	hub_uuid, err := uuid.Parse(hub_id)
	if err != nil {
		log.Println("error when converting to uuid:", hub_id)
		return ""
	}
	h.Hubs[hub_uuid] = hub
	go hub.Run()

	return hub_id
}

func (h *HubManager) GetHub(hub_uuid_string string) *Hub {
	//try to parse string into uuid
	hub_uuid, err := uuid.Parse(hub_uuid_string)
	if err != nil {
		log.Println("error when converting to uuid:", hub_uuid_string)
		return nil
	}
	return h.Hubs[hub_uuid]
}

func (h *HubManager) GetHubListIds() []string {
	hubsIds := make([]string, 0, len(h.Hubs))
	for id := range h.Hubs {
		hubsIds = append(hubsIds, id.String())
	}

	sort.Strings(hubsIds)

	return hubsIds
}

func (h *HubManager) UnregisterFromAll(client *Client) {
	for _, hub := range h.Hubs {
		select {
		case hub.Unregister <- client:
		default:
		}
	}
}

func (h *Hub) DisconnectClient(clientId string) {
	h.UnregisterId <- clientId
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true

			for _, msg := range h.Messages {
				select {
				case client.Send <- []byte(msg.Content):
				default:
					break
				}
			}
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
			}
		case id := <-h.UnregisterId:
			for client := range h.Clients {
				if client.Id == id {
					close(client.Send)
					delete(h.Clients, client)
					break
				}
			}
		case message := <-h.Broadcaster:
			var payload shared.MessagePayload
			if err := json.Unmarshal(message, &payload); err != nil {
				log.Println("Error when extracting message payload: ", err)
				continue
			}

			for client := range h.Clients {
				select {
				case client.Send <- message:
					continue
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (h *Hub) getClientbyId(clientId string) *Client {
	for client := range h.Clients {
		if client.Id == clientId {
			return client
		}
	}

	return nil
}
