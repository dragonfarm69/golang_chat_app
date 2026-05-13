package data

import (
	"bytes"
	shared "chat-app-server/Shared"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (db *DataStorage) CreateUserOnDB(ctx context.Context, user shared.RegisterPayload, userId string) error {
	table := pgx.Identifier{db.schema, "users"}.Sanitize()

	sql := fmt.Sprintf(`
        INSERT INTO %s (id, email, username, avatar_url, created_at, updated_at)
        VALUES (@id, @email, @username, @avatar_url, @created_at, @updated_at)
        RETURNING id
    `, table)

	var id string
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	err := db.Db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
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

func CreateUserOnKeyCloak(ctx context.Context, adminClient *http.Client, user shared.RegisterPayload) (string, error) {
	apiUrl := os.Getenv("KEYCLOAK_API_URL")
	if apiUrl == "" {
		return "", fmt.Errorf("KEYCLOAK_API_URL is not set")
	}

	keycloakPayload := shared.KeyCloakUserPayload{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Enabled:   true,
		Credentials: []shared.CredentialPayload{
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

func (db *DataStorage) CreateNewUser(ctx context.Context, adminClient *http.Client, user shared.RegisterPayload) error {
	//create unique ID for the user.
	UserKeycloakId, err := CreateUserOnKeyCloak(ctx, adminClient, user)
	if err != nil {
		log.Printf("Failed to create new user on Keycloak: %v", err)
		return fmt.Errorf("failed to add new user on Keycloak: %w", err)
	}

	log.Println("User uuid: ", UserKeycloakId)

	//use the keycloak user id for db
	err = db.CreateUserOnDB(ctx, user, UserKeycloakId)

	if err != nil {
		log.Printf("Failed to create user on DB: %v", err)
		return fmt.Errorf("failed to create user on DB: %w", err)
	}

	return nil
}

func (db *DataStorage) FetchUserInfo(ctx context.Context, username string) (shared.UserPayload, error) {
	//check in redis first
	key := fmt.Sprintf("user:%s", username)
	log.Println("Fetching cache data with key: ", key)
	value, err := db.Redis_db.Get(ctx, key).Result()
	if err == nil {
		var userInfo shared.UserPayload
		err = json.Unmarshal([]byte(value), &userInfo)
		if err == nil {
			log.Println("Got user data from redis: ")
			return userInfo, nil
		}

		log.Println("Error when trying to marshal: ", err)
	}

	log.Println("User not found or error: ", err)

	table := pgx.Identifier{db.schema, "users"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT id, username, email, avatar_url, status, created_at, updated_at FROM %s WHERE username = $1
	`, table)

	var user shared.UserModel
	err = db.Db_pool.QueryRow(ctx, sql, username).Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.Avatar_url,
		&user.Status,
		&user.Created_at,
		&user.Updated_at,
	)

	if err != nil {
		return shared.UserPayload{}, fmt.Errorf("failed to fetch user info: %w", err)
	}

	userInfo := shared.UserPayload{
		ID:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.Created_at.Format(time.RFC3339),
	}

	if user.Avatar_url.Valid {
		urlStr := user.Avatar_url.String
		userInfo.AvatarURL = &urlStr
	}

	if user.Updated_at.Valid {
		timeStr := user.Updated_at.Time.Format(time.RFC3339)
		userInfo.UpdatedAt = &timeStr
	}

	//cache user information
	log.Println("User not found - Caching user data")
	db.CacheUserInfoRequest(ctx, userInfo)

	return userInfo, nil
}

func (db *DataStorage) BlacklistToken(ctx context.Context, token string) {
	//1 hour of black list
	err := db.Redis_db.Set(ctx, "blacklist:"+token, "1", 1*time.Hour).Err()
	if err != nil {
		log.Println("Error when trying to blacklist token: ", err)
		return
	}
}

func (db *DataStorage) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := db.Redis_db.Exists(ctx, "blacklist:"+token).Result()
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

func (db *DataStorage) CacheUserInfoRequest(ctx context.Context, user_info shared.UserPayload) {
	key := fmt.Sprintf("user:%s", user_info.Username)
	info, _ := json.Marshal(user_info)
	//TTL: 3 hours
	db.Redis_db.Set(ctx, key, info, 3*time.Hour)
}
