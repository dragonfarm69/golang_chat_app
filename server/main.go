package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/cors"
)

type JwksKey struct {
	Kid string `json: "kid"`
	Kty string `json: "kty"`
	Alg string `json: "alg"`
	Use string `json: "use"`
	N   string `json: "n"`
	E   string `json: "e"`
}

type JwkResponse struct {
	Keys []JwksKey `json: "keys"`
}

var addr = flag.String("addr", ":8080", "chat server service")
var public_keys keyfunc.Keyfunc

func fetch_publick_keys() {
	// {keycloak-server}/realms/{realm-name}/protocol/openid-connect/certs
	var URL = getPublicKeyEndpoint()

	var err error
	public_keys, err = keyfunc.NewDefault([]string{URL})
	if err != nil {
		log.Fatalf("Failed to create a keyfunc.Keyfunc from the server's URL.\nError: %s", err)
	}
}

func is_valid_token(token string) bool {
	status, err := jwt.Parse(token, public_keys.Keyfunc)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			fmt.Println("Error: The string provided is not a valid JWT format.")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			fmt.Println("Error: The signature is invalid (Possible tampering!).")
		case errors.Is(err, jwt.ErrTokenExpired):
			fmt.Println("Error: The token has naturally expired.")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			fmt.Println("Error: The token is not active yet.")
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			// This catches issues like wrong Issuer or wrong Audience
			fmt.Println("Error: The token claims are invalid.")
		default:
			// This will catch jwks.Keyfunc errors (like "kid not found")
			// or network errors if it tried to fetch new keys and failed.
			fmt.Printf("Error: Token validation failed: %v\n", err)
		}
		return false
	}

	if status.Valid {
		return true
	}

	return false
}

func main() {
	flag.Parse()
	hubManager := newHubManager()
	// hub := newHub()
	// go hub.run()

	mux := http.NewServeMux()
	fetch_publick_keys()

	hubManager.createNewHub("temp name")
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// hubId := r.URL.Query().Get("hub")
		user_id := r.URL.Query().Get("user_id")
		println("client id: ", user_id)
		if user_id == "" {
			http.Error(w, "Unknown User", http.StatusBadRequest)
			return
		}

		serveWs(hubManager, w, r, user_id)
	})
	mux.HandleFunc("/newhub", func(w http.ResponseWriter, r *http.Request) {
		hub_id := hubManager.createNewHub("temp")
		w.Write([]byte(hub_id))
	})
	mux.HandleFunc("/hublist", func(w http.ResponseWriter, r *http.Request) {
		lists := hubManager.getHubListIds()
		s := ""
		for i, id := range lists {
			if i > 0 {
				s += ","
			}
			s += id
		}
		w.Write([]byte(s))
	})
	mux.HandleFunc("/disconnect/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("diconnecting")
		hubId := r.URL.Query().Get("hub")
		clientId := r.URL.Query().Get("client")
		log.Println(hubId)
		log.Println(clientId)

		if hubId == "" || clientId == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		hub := hubManager.getHub(hubId)
		if hub == nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		hub.disconnectClient(clientId)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
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
		adminClient := Config.Client(ctx)

		log.Println("payload: ", payload)

		if err := createNewUser(ctx, adminClient, payload); err != nil {
			log.Printf("Error creating new user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
	})
	mux.HandleFunc("/fetch_rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user_id := r.URL.Query().Get("user_id")

		if user_id == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		ctx := r.Context()
		rooms, err := fetchRoomsBasedOnUserId(ctx, user_id)
		log.Println(rooms)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to fetch room list", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rooms)
	})
	mux.HandleFunc("/fetch_room_message", func(w http.ResponseWriter, r *http.Request) {
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
		rooms, err := fetchRoomMessage(ctx, payload.Room_id, payload.Offset_id)
		log.Println(rooms)
		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to fetch room list", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rooms)
	})
	mux.HandleFunc("/fetch_user_info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		username := r.URL.Query().Get("username")

		log.Println("Username: ", username)

		if username == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		ctx := r.Context()
		user, err := fetchUserInfo(ctx, username)

		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	})
	mux.HandleFunc("/refreshToken", func(w http.ResponseWriter, r *http.Request) {
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
	})
	mux.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
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

		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token := authHeader[7:]
		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		// // token now contains the bearer token
		log.Println("Bearer token:", token)

		if payload.UserId == "" || payload.RoomId == "" {
			http.Error(w, "user_id and invite_code are required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		err := addUserToRoom(ctx, payload.UserId, payload.RoomId)

		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to join room", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			UserId   string `json:"user_id"`
			RoomName string `json:"room_name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token := authHeader[7:]
		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		isValid := is_valid_token(token)

		if !isValid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if payload.UserId == "" || payload.RoomName == "" {
			http.Error(w, "user_id and room name are required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		err := createNewRoom(ctx, payload.UserId, payload.RoomName)

		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to create room", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	//Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handler := c.Handler(mux)

	err := http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("error when starting server: ", err)
	}
}
