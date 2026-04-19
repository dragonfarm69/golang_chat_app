package main

import (
	"context"
	"errors"
	"flag"
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

type App struct {
	redis_db      *redis.Client
	db_pool       *pgxpool.Pool
	cloud_storage *s3.Client
	hubManager    *HubManager
	config        *clientcredentials.Config
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

	cloud_client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("http://localhost:9000")
		o.UsePathStyle = true
	})
	log.Println("minIO should work now")

	return &App{
		redis_db:      client,
		db_pool:       Pool,
		cloud_storage: cloud_client,
		config:        Config,
		hubManager:    newHubManager(),
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
	mux := http.NewServeMux()
	fetchPublicToken()

	ctx := context.Background()
	app, err := NewApp(ctx)
	if err != nil {
		panic(err)
	}
	defer app.db_pool.Close()
	defer app.redis_db.Close()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// hubId := r.URL.Query().Get("hub")
		user_id := r.URL.Query().Get("user_id")
		println("client id: ", user_id)
		if user_id == "" {
			http.Error(w, "Unknown User", http.StatusBadRequest)
			return
		}

		app.serveWs(app.hubManager, w, r, user_id)
	})
	mux.HandleFunc("/logout", app.HandleLogOut)
	mux.HandleFunc("/register", app.HandleRegister)
	mux.HandleFunc("/refreshToken", app.HandleRefreshToken)
	mux.HandleFunc("/api/room", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/api/disconnect/", app.HandleDisconnect)
	mux.HandleFunc("/api/fetch_room_message", app.HandleFetchMessages)
	mux.HandleFunc("/api/fetch_user_info", app.HandleFetchUserInfo)
	mux.HandleFunc("/api/join", app.HandleJoinRoom)

	//middleware
	mux.Handle("/api", http.StripPrefix("/api", app.TokenValidation(mux)))

	//Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allow all origins
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	handler := c.Handler(mux)

	err = http.ListenAndServe(*addr, handler)
	if err != nil {
		log.Fatal("error when starting server: ", err)
	}
}
