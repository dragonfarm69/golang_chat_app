package main

import (
	handler "chat-app-server/Handler"
	"chat-app-server/Internal/app"
	"chat-app-server/Internal/data"
	"chat-app-server/Internal/storage"
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/cors"
)

type Server struct {
	dataStorage  *data.DataStorage
	cloudService *storage.S3Service
	appLogic     *app.App
}

func InitializeServer(ctx context.Context) (*Server, error) {
	dataStore, err := data.NewDataStorage(ctx, os.Getenv("DB_URL"), os.Getenv("REDIS_URL"), os.Getenv("SCHEMA_NAME"))
	if err != nil {
		return nil, err
	}
	storageService, err := storage.NewCloudService()
	if err != nil {
		return nil, err
	}
	return &Server{
		dataStorage:  dataStore,
		cloudService: storageService,
		appLogic:     app.NewApp(),
	}, nil
}

var addr = flag.String("addr", ":8080", "chat server service")
var public_keys keyfunc.Keyfunc

func fetchPublicToken() {
	var URL = os.Getenv("KEYCLOAK_PUBLIC_KEY_ENDPOINT")

	var err error
	public_keys, err = keyfunc.NewDefault([]string{URL})
	if err != nil {
		log.Fatalf("Failed to create a keyfunc.Keyfunc from the server's URL.\nError: %s", err)
	}
}

func isValidToken(token string) bool {
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

func (server *Server) TokenValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		status := isValidToken(token)

		ctx := r.Context()
		//check if token is blacklisted
		isBlacklisted, err := server.dataStorage.IsTokenBlacklisted(ctx, token)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
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
	mainMux := http.NewServeMux()
	protectedMux := http.NewServeMux()
	fetchPublicToken()

	ctx := context.Background()
	server, err := InitializeServer(ctx)
	if err != nil {
		log.Fatal("Failed to initialize server: ", err)
	}

	// Create handlers with their dependencies
	authHandler := &handler.AuthHandler{
		Storage:    server.dataStorage,
		Config:     server.appLogic.Config,
		HubManager: server.appLogic.HubManager,
	}
	messageHandler := &handler.MessageHandler{
		Storage:    server.dataStorage,
		Service:    server.cloudService,
		HubManager: server.appLogic.HubManager,
	}
	roomHandler := &handler.RoomHandler{
		Storage: server.dataStorage,
	}
	userHandler := &handler.UserHandler{
		Storage: server.dataStorage,
	}

	// Public routes
	mainMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		user_id := r.URL.Query().Get("user_id")
		if user_id == "" {
			http.Error(w, "Unknown User", http.StatusBadRequest)
			return
		}
		app.ServeWs(server.appLogic.HubManager, w, r, user_id, server.dataStorage)
	})
	mainMux.HandleFunc("/auth/logout", authHandler.HandleLogOut)
	mainMux.HandleFunc("/auth/register", authHandler.HandleRegister)
	mainMux.HandleFunc("/auth/refresh_token", authHandler.HandleRefreshToken)
	mainMux.HandleFunc("/service/webhooks/minio", messageHandler.HandleMediaWebhookResponse)

	// Protected routes
	protectedMux.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			roomHandler.HandleListRoom(w, r)
		case http.MethodPost:
			roomHandler.HandleCreateRoom(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			messageHandler.HandleEditMessage(w, r)
		case http.MethodDelete:
			messageHandler.HandleDeleteMessage(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.HandleFetchUserInfo(w, r)
		// TODO: Implement edit and delete user
		// case http.MethodPatch:
		// case http.MethodDelete:
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/media", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			messageHandler.HandleGeneratePresignURL(w, r, "post")
		case http.MethodDelete:
			//TODO: Implement delete image
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/disconnect/", authHandler.HandleDisconnect)
	protectedMux.HandleFunc("/api/fetch_room_message", messageHandler.HandleFetchMessages)
	protectedMux.HandleFunc("/api/join", roomHandler.HandleJoinRoom)

	// Middleware
	mainMux.Handle("/api/", server.TokenValidation(protectedMux))

	// Configure CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	corsHandler := corsMiddleware.Handler(mainMux)

	err = http.ListenAndServe(*addr, corsHandler)
	if err != nil {
		log.Fatal("error when starting server: ", err)
	}
}
