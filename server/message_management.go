package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func addNewMessageToDB(ctx context.Context, message MessagePayload) (string, error) {
	sql := fmt.Sprintf(`
        INSERT INTO %s.messages (id, content, user_id, edited, created_date, edited_date)
        VALUES (@id, @content, @user_id, @edited, @created_date, @edited_date)
        RETURNING id
    `, pgx.Identifier{DBSchema}.Sanitize())

	var id string

	err := Pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":           "test",
		"content":      "testing things",
		"userid":       "testss",
		"edited":       "false",
		"created_date": "30/12",
		"edited_date":  "30.12",
	}).Scan(&id)

	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return "", fmt.Errorf("something is wrong: %v", err)
	}

	return fmt.Sprintf("%v", id), nil
}
