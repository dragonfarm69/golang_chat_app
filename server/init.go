package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/clientcredentials"
)

var Pool *pgxpool.Pool
var Config *clientcredentials.Config
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

func init() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()
	dbURL := os.Getenv("DB_URL")
	log.Printf("Attempting to connect to database: %s", dbURL)

	Pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		panic(err)
	}

	Config = &clientcredentials.Config{
		ClientID:     os.Getenv("ADMIN_ID"),
		ClientSecret: os.Getenv("ADMIN_SECRET"),
		TokenURL:     os.Getenv("KEYCLOAK_TOKEN_URL"),
	}

	log.Println("Environment loaded")
}
