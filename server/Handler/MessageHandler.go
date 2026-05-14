package handler

import (
	"chat-app-server/Internal/app"
	"chat-app-server/Internal/data"
	"chat-app-server/Internal/storage"
	shared "chat-app-server/Shared"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/oklog/ulid/v2"
)

type MessageHandler struct {
	Storage    *data.DataStorage
	Service    *storage.S3Service
	HubManager *app.HubManager
}

func (handler *MessageHandler) BroadCastEvent(message_id, room_id, content, event_type string) {
	payload := map[string]string{
		"message_id": message_id,
		"content":    content,
	}
	// broadcast to all users
	responsePayload := &shared.WsEvent{
		Type:    event_type,
		Room_ID: room_id,
		Payload: payload,
	}
	jsonPayload, err := json.Marshal(responsePayload)
	if err != nil {
		log.Println("Error when marshalling payload: ", err)
	}

	hub := handler.HubManager.GetHub(room_id)
	if hub == nil {
		println("Hub not found")
		return
	}

	hub.Broadcaster <- jsonPayload
}

func (handler *MessageHandler) HandleEditMessage(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		RoomId    string `json:"room_id"`
		MessageId string `json:"message_id"`
		Content   string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if payload.MessageId == "" || payload.Content == "" {
		http.Error(w, "Message id, room id and content must be not empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := handler.Storage.EditMessage(ctx, payload.RoomId, payload.MessageId, payload.Content)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to edit message", http.StatusInternalServerError)
		return
	}

	handler.BroadCastEvent(payload.MessageId, payload.RoomId, payload.Content, "EDIT")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (handler *MessageHandler) HandleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		RoomId    string `json:"room_id"`
		MessageId string `json:"message_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if payload.MessageId == "" || payload.RoomId == "" {
		http.Error(w, "Message id and room id must be not empty", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err := handler.Storage.DeleteMessage(ctx, payload.RoomId, payload.MessageId)

	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	handler.BroadCastEvent(payload.MessageId, payload.RoomId, "", "DELETE")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (handler *MessageHandler) HandleFetchMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Room_id   string `json:"room_id"`
		Offset_id string `json:"offset_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if payload.Room_id == "" {
		http.Error(w, "Room id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	rooms, err := handler.Storage.FetchRoomMessage(ctx, payload.Room_id, payload.Offset_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to fetch room list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

func (handler *MessageHandler) HandleGeneratePresignURL(w http.ResponseWriter, r *http.Request, url_type string) {
	var payload struct {
		Files   []shared.FileMetaData `json:"files"`
		Room_ID string                `json:"room_id"`
		User_ID string                `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	var urls []string
	var upload_type string = "chat-media"
	for _, val := range payload.Files {
		var content_type string

		switch val.FileType {
		case "image/png", "image/jpeg", "image/jpg", "image/gif", "image/webp":
			content_type = "image"

		case "video/mp4", "video/webm", "video/quicktime":
			content_type = "video"

		default:
			errMsg := fmt.Sprintf("Unsupported file type: %s", val.FileType)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		id := ulid.Make().String()
		uniqueKey := fmt.Sprintf("%s/%s/%s/%s", content_type, payload.Room_ID, id, val.FileName)
		log.Println("Image unique key: ", uniqueKey)
		_, err := handler.Storage.AddNewPendingMediaMessage(ctx, id, content_type, payload.User_ID, payload.Room_ID, uniqueKey)
		if err != nil {
			log.Println("Error when trying to add new pending message: ", err)
			continue
		}

		urlStr, err := handler.Service.GeneratePutPresignedURL(r.Context(), upload_type, uniqueKey, content_type)
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
}

func (handler *MessageHandler) HandleMediaWebhookResponse(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	if len(payload.Records) == 0 {
		http.Error(w, "no records in payload", http.StatusBadRequest)
		return
	}

	key := payload.Records[0].S3.Object.Key
	splitted_strings := strings.Split(key, "%2F")
	message_id := splitted_strings[2]
	room_id := splitted_strings[1]

	// ("%s/%s/%s/%s", content_type, payload.Room_ID, id, val.FileName)
	ctx := r.Context()
	handler.Storage.UpdateMessageState(ctx, message_id, "SENT")
	log.Println("Message id: ", message_id)

	hub := handler.HubManager.GetHub(room_id)
	message, err := handler.Storage.GenerateResponsePayload(ctx, message_id)
	if err != nil {
		log.Println(err)
		return
	}

	jsonPayload, err := json.Marshal(message)
	if err != nil {
		log.Println("Error when marshalling payload: ", err)
		return
	}
	if hub == nil {
		log.Println("Webhook: no active hub for room:", room_id)
		w.WriteHeader(http.StatusOK)
		return
	}
	hub.Broadcaster <- jsonPayload
}
