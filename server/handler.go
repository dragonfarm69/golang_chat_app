package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (app *App) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
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
	err := app.createNewRoom(ctx, payload.UserId, payload.RoomName)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to create room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (app *App) HandleListRoom(w http.ResponseWriter, r *http.Request) {
	user_id := r.URL.Query().Get("user_id")

	if user_id == "" {
		http.Error(w, "Can't be empty", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	rooms, err := app.fetchRoomsBasedOnUserId(ctx, user_id)
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

func (app *App) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload RegisterPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	adminClient := app.config.Client(ctx)

	log.Println("payload: ", payload)

	if err := app.createNewUser(ctx, adminClient, payload); err != nil {
		log.Printf("Error creating new user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func (app *App) HandleFetchMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Room_id   string `json:"room_id"`
		Offset_id string `json:"offset_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if payload.Room_id == "" {
		http.Error(w, "Room id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rooms, err := app.fetchRoomMessage(ctx, payload.Room_id, payload.Offset_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch room list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

func (app *App) HandleFetchUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")

	if username == "" {
		http.Error(w, "Can't be empty", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	user, err := app.fetchUserInfo(ctx, username)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (app *App) HandleJoinRoom(w http.ResponseWriter, r *http.Request) {
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
	err := app.addUserToRoom(ctx, payload.UserId, payload.RoomId)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (app *App) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if payload.RefreshToken == "" {
		http.Error(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	refreshUserToken(ctx, payload.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode()
}

func (app *App) HandleLogOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	header := r.Header.Get("Authorization")
	if header == "" {
		http.Error(w, "Invalid header", http.StatusBadRequest)
		return
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization format", http.StatusBadRequest)
		return
	}
	token := parts[1]

	app.blacklistToken(ctx, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("User logged out successfully")
}

func (app *App) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	log.Print("diconnecting")
	hubId := r.URL.Query().Get("hub")
	clientId := r.URL.Query().Get("client")
	log.Println(hubId)
	log.Println(clientId)

	if hubId == "" || clientId == "" {
		http.Error(w, "Can't be empty", http.StatusNotFound)
		return
	}

	hub := app.hubManager.getHub(hubId)
	if hub == nil {
		http.Error(w, "404 Not found", http.StatusNotFound)
		return
	}

	hub.disconnectClient(clientId)
}
