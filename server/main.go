package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
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

type App struct {
	redis_db *redis.Client
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
			log.Println("Error: The string provided is not a valid JWT format.")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			log.Println("Error: The signature is invalid (Possible tampering!).")
		case errors.Is(err, jwt.ErrTokenExpired):
			log.Println("Error: The token has naturally expired.")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			log.Println("Error: The token is not active yet.")
		case errors.Is(err, jwt.ErrTokenInvalidClaims):
			// This catches issues like wrong Issuer or wrong Audience
			log.Println("Error: The token claims are invalid.")
		default:
			// This will catch jwks.Keyfunc errors (like "kid not found")
			// or network errors if it tried to fetch new keys and failed.
			log.Printf("Error: Token validation failed: %v\n", err)
		}
		return false
	}

	if status.Valid {
		return true
	}

	return false
}

func (app *App) TokenValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Invalid header", http.StatusBadRequest)
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusBadRequest)
		}
		token := parts[1]

		status := is_valid_token(token)

		ctx := r.Context()
		//check if token is blacklisted
		isBlacklisted, err := app.is_token_blacklisted(ctx, token)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		if isBlacklisted {
			http.Error(w, "Unauthorized Token", http.StatusUnauthorized)
			return
		}

		if status {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Error when trying to serve http", http.StatusInternalServerError)
			return
		}
	})
}

func main() {
	flag.Parse()
	hubManager := newHubManager()
	// hub := newHub()
	// go hub.run()

	mux := http.NewServeMux()
	fetch_publick_keys()

	//redis connection
	//TODO: SET UP ENV
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // use default DB
		Protocol: 2,
	})
	defer rdb.Close()

	app := &App{
		redis_db: rdb,
	}

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
	mux.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		//get room list
		case http.MethodGet:
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
		//create room
		case http.MethodPost:
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
		//delete room
		// case http.MethodDelete:

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/hublist", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Invalid header", http.StatusBadRequest)
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusBadRequest)
		}
		token := parts[1]

		app.blacklist_token(ctx, token)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("User logged out successfully")
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
	mux.HandleFunc("/api/fetch_room_message", func(w http.ResponseWriter, r *http.Request) {
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
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(rooms)
	})
	mux.HandleFunc("/api/fetch_user_info", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/api/join", func(w http.ResponseWriter, r *http.Request) {
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
		err := addUserToRoom(ctx, payload.UserId, payload.RoomId)

		if err != nil {
			log.Println(err)
			http.Error(w, "Failed to join room", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	//middleware
	mux.Handle("/api", http.StripPrefix("/api", app.TokenValidation(mux)))

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
