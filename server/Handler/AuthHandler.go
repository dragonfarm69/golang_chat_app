package handler

import (
	auth "chat-app-server/Internal/Auth"
	"chat-app-server/Internal/app"
	"chat-app-server/Internal/data"
	shared "chat-app-server/Shared"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2/clientcredentials"
)

type AuthHandler struct {
	Storage    *data.DataStorage
	Config     *clientcredentials.Config
	HubManager *app.HubManager
}

func (handler *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload shared.RegisterPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	adminClient := handler.Config.Client(ctx)

	log.Println("payload: ", payload)

	if err := handler.Storage.CreateNewUser(ctx, adminClient, payload); err != nil {
		log.Printf("Error creating new user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func (handler *AuthHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
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
	// (ctx context.Context, refresh_token string, refresh_token_endpoint, clientID, clientSecret string)
	// TODO: ADD GETENV
	auth.RefreshUserToken(ctx, payload.RefreshToken, handler.Config.TokenURL, handler.Config.ClientID, handler.Config.ClientSecret)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode()
}

func (handler *AuthHandler) HandleLogOut(w http.ResponseWriter, r *http.Request) {
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

	handler.Storage.BlacklistToken(ctx, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("User logged out successfully")
}

func (handler *AuthHandler) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	log.Print("diconnecting")
	hubId := r.URL.Query().Get("hub")
	clientId := r.URL.Query().Get("client")
	log.Println(hubId)
	log.Println(clientId)

	if hubId == "" || clientId == "" {
		http.Error(w, "Can't be empty", http.StatusNotFound)
		return
	}

	hub := handler.HubManager.GetHub(hubId)
	if hub == nil {
		http.Error(w, "404 Not found", http.StatusNotFound)
		return
	}

	hub.DisconnectClient(clientId)
}
