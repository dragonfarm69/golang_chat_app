package handler

import (
	"chat-app-server/Internal/data"
	"encoding/json"
	"log"
	"net/http"
)

type UserHandler struct {
	Storage *data.DataStorage
}

func (handler *UserHandler) HandleFetchUserInfo(w http.ResponseWriter, r *http.Request) {
	// log.Println("hello")
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.URL.Query().Get("username")

	if username == "" {
		http.Error(w, "Can't be empty", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	user, err := handler.Storage.FetchUserInfo(ctx, username)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}
