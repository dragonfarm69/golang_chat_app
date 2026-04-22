package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type WsEvent struct {
	Type    string `json:"action"`
	Room_ID string `json:"room_id"`
	Payload any    `json:"payload"`
}

func (app *App) addNewMessageToDB(ctx context.Context, message MessagePayload) (string, error) {
	sql := fmt.Sprintf(`
        INSERT INTO %s.messages (id, content, user_id, room_id, created_at, updated_at)
        VALUES (@id, @content, @user_id, @room_id, @created_at, @updated_at)
        RETURNING id
    `, pgx.Identifier{DBSchema}.Sanitize())

	var id string
	createdAt, _ := time.Parse(time.RFC3339, message.TimeStamp)

	err := app.db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":         message.Id,
		"content":    message.Content,
		"room_id":    message.Room_ID,
		"user_id":    message.User_ID,
		"created_at": createdAt,
		"updated_at": nil,
	}).Scan(&id)

	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return "", fmt.Errorf("something is wrong: %v", err)
	}

	return fmt.Sprintf("%v", id), nil
}

func (app *App) addNewMessageToRedis(ctx context.Context, message MessagePayload) {
	roomMessage := RoomMessage{
		Id:         message.Id,
		Owner_name: message.UserName,
		Room_ID:    message.Room_ID,
		Content:    message.Content,
		TimeStamp:  message.TimeStamp,
	}

	key := fmt.Sprintf("room:%s:recent_messages", roomMessage.Room_ID)
	msg, _ := json.Marshal(roomMessage)
	app.redis_db.LPush(ctx, key, msg)

	//Make sure to keep only 50 newest message
	app.redis_db.LTrim(ctx, key, 0, 49)
}

func (app *App) editMessage(ctx context.Context, room_id string, message_id string, content string) error {
	//update in db
	sql := fmt.Sprintf(`
        UPDATE %s.messages 
        SET content = @content, updated_at = @updated_at
        WHERE id = @id
        RETURNING id
    `, pgx.Identifier{DBSchema}.Sanitize())

	var id string
	updated_at := time.Now()

	err := app.db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":         message_id,
		"content":    content,
		"updated_at": updated_at,
	}).Scan(&id)

	if err != nil {
		log.Println("SOMETHING IS WRONG WHEN TRYING TO EDIT MESSAGE: ", err)
		return err
	}

	log.Println("Edited message, broadcasting to all user")

	//delete redis
	key := fmt.Sprintf("room:%s:recent_messages", room_id)
	app.redis_db.Del(ctx, key)

	payload := map[string]string{
		"message_id": message_id,
		"content":    content,
	}

	//broadcast to all users
	responsePayload := &WsEvent{
		Type:    "EDIT",
		Room_ID: room_id,
		Payload: payload,
	}
	jsonPayload, err := json.Marshal(responsePayload)
	if err != nil {
		log.Println("Error when marshalling payload: ", err)
		return err
	}
	hub := app.hubManager.getHub(room_id)
	hub.broadcaster <- jsonPayload
	return nil
}

func (app *App) deleteMessage(ctx context.Context, room_id string, message_id string) error {
	//update in db
	sql := fmt.Sprintf(`
        DELETE FROM %s.messages 
        WHERE id = @id
    `, pgx.Identifier{DBSchema}.Sanitize())

	_, err := app.db_pool.Exec(ctx, sql, pgx.NamedArgs{
		"id": message_id,
	})

	if err != nil {
		log.Println("SOMETHING IS WRONG WHEN TRYING TO DELETE MESSAGE: ", err)
		return err
	}

	log.Println("Deleted message, broadcasting to all user")

	//delete redis
	key := fmt.Sprintf("room:%s:recent_messages", room_id)
	app.redis_db.Del(ctx, key)

	payload := map[string]string{
		"message_id": message_id,
	}

	//broadcast to all users
	responsePayload := &WsEvent{
		Type:    "DELETE",
		Room_ID: room_id,
		Payload: payload,
	}
	jsonPayload, err := json.Marshal(responsePayload)
	if err != nil {
		log.Println("Error when marshalling payload: ", err)
		return err
	}
	hub := app.hubManager.getHub(room_id)
	hub.broadcaster <- jsonPayload

	return nil
}
