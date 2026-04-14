package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var DBSchema string

func getKeyCloakLoginURL() string {
	return os.Getenv("KEYCLOAK_ACCOUNT_LOGIN_URL")
}

func getAdminID() string {
	return os.Getenv("ADMIN_ID")
}

func getAdminSecret() string {
	return os.Getenv("ADMIN_SECRET")
}

func getClientID() string {
	return os.Getenv("CLIENT_ID")
}

func getClientSecret() string {
	return os.Getenv("CLIENT_SECRET")
}

func getKeyCloakTokenURL() string {
	return os.Getenv("KEYCLOAK_TOKEN_URL")
}

func getKeyCloakAPIURL() string {
	return os.Getenv("KEYCLOAK_API_URL")
}

func getDBURL() string {
	return os.Getenv("DB_URL")
}

func getSchemaName() string {
	return os.Getenv("SCHEMA_NAME")
}

func getRefreshTokenEndpoint() string {
	return os.Getenv("REFRESH_TOKEN_ENDPOINT")
}

func getPublicKeyEndpoint() string {
	return os.Getenv("KEYCLOAK_PUBLIC_KEY_ENDPOINT")
}

func init() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	DBSchema = getSchemaName()
	log.Println("Environment loaded")
}
