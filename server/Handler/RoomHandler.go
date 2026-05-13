package handler

import (
	"chat-app-server/Internal/data"
	"encoding/json"
	"log"
	"net/http"
)

type RoomHandler struct {
	Storage *data.DataStorage
}

func (handler *RoomHandler) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserId   string `json:"user_id"`
		RoomName string `json:"room_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if payload.UserId == "" || payload.RoomName == "" {
		http.Error(w, "user_id and room name are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := handler.Storage.CreateNewRoom(ctx, payload.UserId, payload.RoomName)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (handler *RoomHandler) HandleListRoom(w http.ResponseWriter, r *http.Request) {
	user_id := r.URL.Query().Get("user_id")

	if user_id == "" {
		http.Error(w, "Can't be empty", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	rooms, err := handler.Storage.FetchRoomsBasedOnUserId(ctx, user_id)
	log.Println(rooms)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch room list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

func (handler *RoomHandler) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		UserId string `json:"user_id"`
		RoomId string `json:"invite_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if payload.UserId == "" || payload.RoomId == "" {
		http.Error(w, "user_id and invite_code are required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := handler.Storage.AddUserToRoom(ctx, payload.UserId, payload.RoomId)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
