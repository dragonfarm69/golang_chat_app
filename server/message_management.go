package main

import (
	"context"
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
