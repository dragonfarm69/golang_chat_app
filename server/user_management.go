package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type CredentialPayload struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

// UserPayload for keycloak
type KeyCloakUserPayload struct {
	FirstName   string              `json:"firstName"`
	LastName    string              `json:"lastName"`
	Email       string              `json:"email"`
	Enabled     bool                `json:"enabled"`
	Credentials []CredentialPayload `json:"credentials"`
}

type RegisterPayload struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// UserModel for the database.
type UserModel struct {
	id         string
	username   string
	email      string
	avatar_url sql.NullString
	status     string
	created_at time.Time
	updated_at sql.NullTime
}

// use this for response to requests
type UserPayload struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url"` // Pointer = JSON null if nil
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at"` // Pointer for optional dates
}

func createUserOnDB(ctx context.Context, user RegisterPayload, userId string) error {
	// Ensure schema is not empty, default to "public" if not set.
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	table := pgx.Identifier{schema, "users"}.Sanitize()

	sql := fmt.Sprintf(`
        INSERT INTO %s (id, email, username, avatar_url, created_at, updated_at)
        VALUES (@id, @email, @username, @avatar_url, @created_at, @updated_at)
        RETURNING id
    `, table)

	var id string
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	err := Pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":         userId,
		"username":   fullName,
		"email":      user.Email,
		"avatar_url": nil,
		"created_at": time.Now(),
		"updated_at": nil,
	}).Scan(&id)

	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return fmt.Errorf("something is wrong: %v", err)
	}

	return nil
}

func createUserOnKeyCloak(ctx context.Context, adminClient *http.Client, user RegisterPayload) (string, error) {
	apiUrl := os.Getenv("KEYCLOAK_API_URL")
	if apiUrl == "" {
		return "", fmt.Errorf("KEYCLOAK_API_URL is not set")
	}

	keycloakPayload := KeyCloakUserPayload{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Enabled:   true,
		Credentials: []CredentialPayload{
			{
				Type:      "password",
				Value:     user.Password,
				Temporary: false,
			},
		},
	}

	payload, err := json.Marshal(keycloakPayload)
	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := adminClient.Do(req)
	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return "", fmt.Errorf("failed when requesting: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("user creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	userId := strings.TrimPrefix(resp.Header.Get("Location"), apiUrl+"/")
	return userId, nil
}

func createNewUser(ctx context.Context, adminClient *http.Client, user RegisterPayload) error {
	//create unique ID for the user.
	UserKeycloakId, err := createUserOnKeyCloak(ctx, adminClient, user)
	if err != nil {
		log.Printf("Failed to create new user on Keycloak: %v", err)
		return fmt.Errorf("failed to add new user on Keycloak: %w", err)
	}

	log.Println("User uuid: ", UserKeycloakId)

	//use the keycloak user id for db
	err = createUserOnDB(ctx, user, UserKeycloakId)

	if err != nil {
		log.Printf("Failed to create user on DB: %v", err)
		return fmt.Errorf("failed to create user on DB: %w", err)
	}

	return nil
}

func fetchUserInfo(ctx context.Context, username string) (UserPayload, error) {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	table := pgx.Identifier{schema, "users"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT id, username, email, avatar_url, status, created_at, updated_at FROM %s WHERE email = $1
	`, table)

	var user UserModel
	err := Pool.QueryRow(ctx, sql, username).Scan(
		&user.id,
		&user.username,
		&user.email,
		&user.avatar_url,
		&user.status,
		&user.created_at,
		&user.updated_at,
	)

	if err != nil {
		return UserPayload{}, fmt.Errorf("failed to fetch user info: %w", err)
	}

	userInfo := UserPayload{
		ID:        user.id,
		Username:  user.username,
		Email:     user.email,
		Status:    user.status,
		CreatedAt: user.created_at.Format(time.RFC3339),
	}

	if user.avatar_url.Valid {
		urlStr := user.avatar_url.String
		userInfo.AvatarURL = &urlStr
	}

	if user.updated_at.Valid {
		timeStr := user.updated_at.Time.Format(time.RFC3339)
		userInfo.UpdatedAt = &timeStr
	}

	return userInfo, nil
}

func refreshUserToken(ctx context.Context, refresh_token string) {
	refresh_token_endpoint := getRefreshTokenEndpoint()

	clientID := getClientID()
	clientSecret := getClientSecret()

	log.Println(clientID)
	log.Println(clientSecret)

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refresh_token)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, refresh_token_endpoint, strings.NewReader(data.Encode()))

	if err != nil {
		// return "", fmt.Errorf("Failed to create request: %w", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// cookie := &http.Cookie{
	// 	Name:  "refresh_token",
	// 	Value: refresh_token,
	// }
	// req.AddCookie(cookie)
	// req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		// return "", fmt.Errorf("Failed to request token refresh: %w", err)
		return
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	log.Println("response status: ", res.Status)
	log.Println("Data: ", string(body))

	// return string(body)
}

func (app *App) blacklist_token(ctx context.Context, token string) {
	//1 hour of black list
	err := app.redis_db.Set(ctx, "blacklist:"+token, "1", 1*time.Hour).Err()
	if err != nil {
		log.Println("Error when trying to blacklist token: ", err)
		return
	}
}

func (app *App) is_token_blacklisted(ctx context.Context, token string) (bool, error) {
	val, err := app.redis_db.Exists(ctx, "blacklist:"+token).Result()
	if err != nil {
		log.Println("Error when trying to search for token: ", err)
		return false, fmt.Errorf("Error when searching token: %w", err)
	}

	// 1 if exists
	if val == 1 {
		return true, nil
	} else {
		return false, nil
	}
}
