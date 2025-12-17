package main

import (
	"bytes"
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

type User struct {
	ID       string
	Username string
	Email    string
}

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
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
	Password  string `json:"password"`
}

// UserModel for the database.
type UserModel struct {
	ID         string
	Name       string
	Username   string
	Email      string
	ProfileURL string
}

func createUserOnDB(ctx context.Context, user RegisterPayload, userId string) error {
	// Ensure schema is not empty, default to "public" if not set.
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	table := pgx.Identifier{schema, "userinfo"}.Sanitize()

	// Corrected SQL statement to match the provided arguments.
	// Assuming your table has a `created_date` column. If not, remove it.
	sql := fmt.Sprintf(`
        INSERT INTO %s (id, email, name, profile_url, created_date)
        VALUES (@id, @email, @name, @profile_url, @created_date)
        RETURNING id
    `, table)

	var id string
	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	err := Pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":           userId,
		"name":         fullName,
		"email":        user.Email,
		"created_date": time.Now(),
		"profile_url":  "test",
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
		Enabled:   user.Enabled,
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
