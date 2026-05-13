package data

import (
	shared "chat-app-server/Shared"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

func (db *DataStorage) AddNewMessageToDB(ctx context.Context, message shared.MessagePayload) (string, error) {
	sql := fmt.Sprintf(`
        INSERT INTO %s.messages (id, content, user_id, room_id, created_at, updated_at)
        VALUES (@id, @content, @user_id, @room_id, @created_at, @updated_at)
        RETURNING id
    `, pgx.Identifier{db.schema}.Sanitize())

	var id string
	createdAt, _ := time.Parse(time.RFC3339, message.TimeStamp)

	err := db.Db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
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

func (db *DataStorage) AddNewMessageToRedis(ctx context.Context, message shared.MessagePayload) {
	roomMessage := shared.RoomMessage{
		Id:         message.Id,
		Owner_name: message.UserName,
		Room_ID:    message.Room_ID,
		Content:    message.Content,
		TimeStamp:  message.TimeStamp,
	}

	key := fmt.Sprintf("room:%s:recent_messages", roomMessage.Room_ID)
	msg, _ := json.Marshal(roomMessage)
	db.Redis_db.LPush(ctx, key, msg)

	//Make sure to keep only 50 newest message
	db.Redis_db.LTrim(ctx, key, 0, 49)
}

func (db *DataStorage) EditMessage(ctx context.Context, room_id string, message_id string, content string) error {
	//update in db
	sql := fmt.Sprintf(`
        UPDATE %s.messages 
        SET content = @content, updated_at = @updated_at
        WHERE id = @id
        RETURNING id
    `, pgx.Identifier{db.schema}.Sanitize())

	var id string
	updated_at := time.Now()

	err := db.Db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
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
	db.Redis_db.Del(ctx, key)

	//TODO: MAKE MORE SENSE TO MOVE THIS TO WS HANDLER SO THAT WE CAN BROADCAST THE MESSAGE
	// payload := map[string]string{
	// 	"message_id": message_id,
	// 	"content":    content,
	// }
	//broadcast to all users
	// responsePayload := &WsEvent{
	// 	Type:    "EDIT",
	// 	Room_ID: room_id,
	// 	Payload: payload,
	// }
	// jsonPayload, err := json.Marshal(responsePayload)
	// if err != nil {
	// 	log.Println("Error when marshalling payload: ", err)
	// 	return err
	// }

	// hub := hubManager.getHub(room_id)
	// hub.broadcaster <- jsonPayload
	return nil
}

func (db *DataStorage) DeleteMessage(ctx context.Context, room_id string, message_id string) error {
	//update in db
	sql := fmt.Sprintf(`
        DELETE FROM %s.messages 
        WHERE id = @id
    `, pgx.Identifier{db.schema}.Sanitize())

	_, err := db.Db_pool.Exec(ctx, sql, pgx.NamedArgs{
		"id": message_id,
	})

	if err != nil {
		log.Println("SOMETHING IS WRONG WHEN TRYING TO DELETE MESSAGE: ", err)
		return err
	}

	log.Println("Deleted message, broadcasting to all user")

	//delete redis
	key := fmt.Sprintf("room:%s:recent_messages", room_id)
	db.Redis_db.Del(ctx, key)

	//TODO: MAKE MORE SENSE TO MOVE THIS TO WS HANDLER SO THAT WE CAN BROADCAST THE MESSAGE

	// payload := map[string]string{
	// 	"message_id": message_id,
	// }

	// //broadcast to all users
	// responsePayload := &WsEvent{
	// 	Type:    "DELETE",
	// 	Room_ID: room_id,
	// 	Payload: payload,
	// }
	// jsonPayload, err := json.Marshal(responsePayload)
	// if err != nil {
	// 	log.Println("Error when marshalling payload: ", err)
	// 	return err
	// }
	// hub := db.hubManager.getHub(room_id)
	// hub.broadcaster <- jsonPayload

	return nil
}

func (db *DataStorage) AddNewPendingMediaMessage(ctx context.Context, message_id string, message_type string, user_id string, room_id string, key string) (string, error) {
	sql := fmt.Sprintf(`
        INSERT INTO %s.messages (id, content, message_type, user_id, room_id)
        VALUES (@id, @content, @message_type, @user_id, @room_id)
        RETURNING id
    `, pgx.Identifier{db.schema}.Sanitize())

	err := db.Db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":           message_id,
		"content":      key,
		"room_id":      room_id,
		"user_id":      user_id,
		"message_type": message_type,
	}).Scan(&message_id)

	if err != nil {
		log.Println("SOMETHING IS WRONG: ", err)
		return "", fmt.Errorf("something is wrong: %v", err)
	}

	return fmt.Sprintf("%v", message_id), nil
}

func (db *DataStorage) UpdateMessageState(ctx context.Context, message_id string, state string) (string, error) {
	sql := fmt.Sprintf(`
        UPDATE %s.messages 
        SET state = @state
        WHERE id = @id
        RETURNING id
    `, pgx.Identifier{db.schema}.Sanitize())

	id := ulid.Make().String()

	err := db.Db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"id":    message_id,
		"state": state,
	}).Scan(&id)

	if err != nil {
		log.Println("SOMETHING IS WRONG WHEN TRYING TO UPDATE MESSAGE STATE: ", err)
		return "", err
	}

	return fmt.Sprintf("%v", id), nil
}

func (db *DataStorage) GenerateResponsePayload(ctx context.Context, message_id string) (shared.ResponseMessagePayload, error) {
	messagesTable := pgx.Identifier{db.schema, "messages"}.Sanitize()
	usersTable := pgx.Identifier{db.schema, "users"}.Sanitize()
	sql := fmt.Sprintf(`
		SELECT m.id, u.id, COALESCE(u.username, 'Unknown User'), m.room_id, m.content, m.created_at, m.message_type
		FROM %s m
		LEFT JOIN %s u ON m.user_id = u.id
		WHERE m.id = $1
	`, messagesTable, usersTable)
	args := []interface{}{message_id}
	var createdAt time.Time

	var message shared.ResponseMessagePayload
	err := db.Db_pool.QueryRow(ctx, sql, args...).Scan(
		&message.Id,
		&message.User_ID,
		&message.UserName,
		&message.Room_ID,
		&message.Content,
		&createdAt,
		&message.Message_Type,
	)
	if err != nil {
		return shared.ResponseMessagePayload{}, fmt.Errorf("Query failed: %w", err)
	}

	message.OriginalId = message_id
	message.TimeStamp = createdAt.Format(time.RFC3339)
	message.Action = "SEND"

	log.Println("MESSAGES INFO: ", message)
	return message, nil
}
