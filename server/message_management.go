package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func addNewMessageToDB(ctx context.Context, message MessagePayload) (string, error) {
	schema := "chat"
	sql := fmt.Sprintf(`
        INSERT INTO %s.messages (id, content, user_id, room_id, created_at, updated_at)
        VALUES (@id, @content, @user_id, @room_id, @created_at, @updated_at)
        RETURNING id
    `, pgx.Identifier{schema}.Sanitize())

	var id string
	createdAt, _ := time.Parse(time.RFC3339, message.TimeStamp)

	err := Pool.QueryRow(ctx, sql, pgx.NamedArgs{
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
