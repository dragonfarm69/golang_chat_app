package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/rs/cors"
)

var addr = flag.String("addr", ":8080", "chat server service")

func serveMainHtml(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func main() {
	flag.Parse()
	hubManager := newHubManager()
	// hub := newHub()
	// go hub.run()

	mux := http.NewServeMux()

	hubManager.createNewHub("temp name")
	mux.HandleFunc("/", serveMainHtml)
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

	mux.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		hubId := r.URL.Query().Get("hub")
		clientId := r.URL.Query().Get("client")

		if hubId == "" || clientId == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		hub := hubManager.getHub(hubId)
		if hub == nil {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}

		is_client_exists := hub.isClientExists(clientId)
		if is_client_exists {
			w.WriteHeader(http.StatusExpectationFailed)
			http.Error(w, "Client already exist !", http.StatusBadRequest)
			return
		}

		// hub.disconnectClient(clientId)
		// serveWs(hub, w, r)
		w.WriteHeader(http.StatusOK)
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

		payload.Enabled = true

		ctx := r.Context()
		adminClient := Config.Client(ctx)

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
		room_id := r.URL.Query().Get("room_id")

		if room_id == "" {
			http.Error(w, "Can't be empty", http.StatusNotFound)
			return
		}

		ctx := r.Context()
		rooms, err := fetchRoomMessage(ctx, room_id)
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
