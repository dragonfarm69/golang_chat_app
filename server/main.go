package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"golang.org/x/oauth2/clientcredentials"
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

type FileMetaData struct {
	FileName string `json:"file_name"`
	FileSize string `json:"file_size"`
	FileType string `json:"file_type"`
}

type App struct {
	redis_db            *redis.Client
	db_pool             *pgxpool.Pool
	s3_client           *s3.Client
	s3_presigned_client *s3.PresignClient
	hubManager          *HubManager
	config              *clientcredentials.Config
}

func NewApp(ctx context.Context) (*App, error) {
	//load env
	dbURL := os.Getenv("DB_URL")

	//load DB
	log.Printf("Attempting to connect to database: %s", dbURL)
	Pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	//load config
	Config := &clientcredentials.Config{
		ClientID:     os.Getenv("ADMIN_ID"),
		ClientSecret: os.Getenv("ADMIN_SECRET"),
		TokenURL:     os.Getenv("KEYCLOAK_TOKEN_URL"),
	}

	//load redis
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}

	log.Printf("Attempting to connect to database: %s", os.Getenv("REDIS_URL"))
	client := redis.NewClient(opt)

	//load cloud storage
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("minioadmin", "miniocloud", "")),
	)

	if err != nil {
		log.Fatal("cannot load sdk config: ", err)
	}

	s3_client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://localhost:9000")
		o.UsePathStyle = true
	})

	presignedClient := s3.NewPresignClient(s3_client)

	log.Println("minIO should work now")

	return &App{
		redis_db:            client,
		db_pool:             Pool,
		s3_client:           s3_client,
		s3_presigned_client: presignedClient,
		config:              Config,
		hubManager:          newHubManager(),
	}, nil
}

var addr = flag.String("addr", ":8080", "chat server service")
var public_keys keyfunc.Keyfunc

func fetchPublicToken() {
	var URL = getPublicKeyEndpoint()

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
			return
		}
		token := parts[1]

		status := isValidToken(token)

		ctx := r.Context()
		//check if token is blacklisted
		isBlacklisted, err := app.isTokenBlacklisted(ctx, token)
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
	app, err := NewApp(ctx)
	if err != nil {
		panic(err)
	}
	defer app.db_pool.Close()
	defer app.redis_db.Close()

	mainMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// hubId := r.URL.Query().Get("hub")
		user_id := r.URL.Query().Get("user_id")
		println("client id: ", user_id)
		if user_id == "" {
			http.Error(w, "Unknown User", http.StatusBadRequest)
			return
		}

		app.serveWs(app.hubManager, w, r, user_id)
	})
	mainMux.HandleFunc("/auth/logout", app.HandleLogOut)
	mainMux.HandleFunc("/auth/register", app.HandleRegister)
	mainMux.HandleFunc("/auth/refresh_token", app.HandleRefreshToken)
	mainMux.HandleFunc("/service/webhooks/minio", func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "Unauthorized Request", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 && parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusBadRequest)
			return
		}

		secret_token := os.Getenv("MINIO_SECRET_TOKEN")
		token := parts[1]

		if secret_token != token {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}

		type MinioWebhookPayload struct {
			EventName string `json:"EventName"`
			Key       string `json:"Key"`
			Records   []struct {
				EventTime string `json:"eventTime"`
				S3        struct {
					Bucket struct {
						Name string `json:"name"`
					} `json:"bucket"`
					Object struct {
						Key         string `json:"key"`
						Size        int64  `json:"size"`
						ContentType string `json:"contentType"`
					} `json:"object"`
				} `json:"s3"`
			} `json:"Records"`
		}

		var payload MinioWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		}
		key := payload.Records[0].S3.Object.Key

		message_id := strings.Split(key, "%2F")[1]
		ctx := r.Context()
		app.updateMessageState(ctx, message_id, "SENT")
		log.Println("Message id: ", message_id)
	})

	protectedMux.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		//get room list
		case http.MethodGet:
			app.HandleListRoom(w, r)
			//create room
		case http.MethodPost:
			app.HandleCreateRoom(w, r)
		//delete room
		// case http.MethodDelete:

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/message", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		//edit
		case http.MethodPatch:
			app.HandleEditMessage(w, r)
		//delete
		case http.MethodDelete:
			app.HandleDeleteMessage(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.HandleFetchUserInfo(w, r)
		//edit
		case http.MethodPatch:
			app.HandleEditUser(w, r)
		//delete
		case http.MethodDelete:
			app.HandleDeleteUser(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/media", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var payload struct {
				Files   []FileMetaData `json:"files"`
				Room_ID string         `json:"room_id"`
				User_ID string         `json:"user_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid json body", http.StatusBadRequest)
				return
			}

			var urls []string
			for _, val := range payload.Files {
				var upload_type string
				var content_type string

				switch val.FileType {
				case "image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp":
					upload_type = "chat-image"
					content_type = "image"

				case "video/mp4", "video/webm", "video/quicktime":
					upload_type = "chat-video"
					content_type = "video"

				default:
					errMsg := fmt.Sprintf("Unsupported file type: %s", val.FileType)
					http.Error(w, errMsg, http.StatusBadRequest)
					return
				}

				id := ulid.Make().String()
				uniqueKey := fmt.Sprintf("%s/%s/%s", payload.Room_ID, id, val.FileName)
				log.Println("Image unique key: ", uniqueKey)
				_, err := app.addNewPendingMediaMessage(ctx, id, content_type, payload.User_ID, payload.Room_ID, uniqueKey)
				if err != nil {
					log.Println("Error when trying to add new pending message: ", err)
					continue
				}

				urlStr, err := app.generatePutPresignedURL(r.Context(), upload_type, uniqueKey, content_type)
				if err != nil {
					log.Printf("MinIO Presign Error: %v", err)
					http.Error(w, "Failed to create presigned url", http.StatusInternalServerError)
					return
				}

				urls = append(urls, urlStr)
			}
			log.Println("All files: ", payload.Files)
			log.Println("Reuslt URL: ", urls)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(urls)
		case http.MethodDelete:
			var payload struct {
				Files   []FileMetaData `json"files"`
				Room_ID string         `json:"room_id"`
				User_ID string         `json:"user_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid json body", http.StatusBadRequest)
				return
			}

			// log.Println("File name: ", payload.File_name)
			// val, err := app.generateGetPresignedURL(r.Context(), payload.File_name, payload.File_size)
			if err != nil {
				http.Error(w, "Failed to create presigned url", http.StatusInternalServerError)
			}
			// log.Println(val)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	protectedMux.HandleFunc("/api/disconnect/", app.HandleDisconnect)
	protectedMux.HandleFunc("/api/fetch_room_message", app.HandleFetchMessages)
	protectedMux.HandleFunc("/api/join", app.HandleJoinRoom)

	//middleware
	mainMux.Handle("/api/", app.TokenValidation(protectedMux))

	//Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handler := c.Handler(mainMux)

	err = http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("error when starting server: ", err)
	}
}
