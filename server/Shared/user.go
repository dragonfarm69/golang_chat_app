package shared

import (
	"database/sql"
	"time"
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
	Id         string
	Username   string
	Email      string
	Avatar_url sql.NullString
	Status     string
	Created_at time.Time
	Updated_at sql.NullTime
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
