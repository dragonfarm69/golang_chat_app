package app

import (
	"os"

	"golang.org/x/oauth2/clientcredentials"
)

type App struct {
	HubManager *HubManager
	Config     *clientcredentials.Config
}

func NewApp() *App {
	Config := &clientcredentials.Config{
		ClientID:     os.Getenv("ADMIN_ID"),
		ClientSecret: os.Getenv("ADMIN_SECRET"),
		TokenURL:     os.Getenv("KEYCLOAK_TOKEN_URL"),
	}

	return &App{
		HubManager: newHubManager(),
		Config:     Config,
	}
}
