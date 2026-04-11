package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
)

type RoomLite struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Room struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPrivate   bool      `json:"is_private"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomMessage struct {
	Id         string `json:"id"`
	Owner_name string `json:"owner_name"`
	Room_ID    string `json:"room_id"`
	Content    string `json:"content"`
	TimeStamp  string `json:"timeStamp"`
}

const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func createInviteCode() (string, error) {
	codeLength := 8
	b := make([]byte, codeLength)

	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		b[i] = charset[num.Int64()]
	}

	return fmt.Sprintf("INV-%s", string(b)), nil
}

func fetchFullRoomDataBasedOnRoomId(ctx context.Context, room_id string) (Room, error) {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	table := pgx.Identifier{schema, "rooms"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT * FROM %s WHERE id = $1
	`, table)

	var room Room
	err := Pool.QueryRow(ctx, sql, room_id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.IsPrivate,
		&room.OwnerID,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return Room{}, fmt.Errorf("failed to fetch room_id: %w", err)
	}

	return room, nil
}

func fetchRoomLiteDataBasedOnRoomId(ctx context.Context, room_id string) (RoomLite, error) {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	table := pgx.Identifier{schema, "rooms"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT * FROM %s WHERE id = $1
	`, table)

	var room RoomLite
	err := Pool.QueryRow(ctx, sql, room_id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return RoomLite{}, fmt.Errorf("failed to fetch room_id: %w", err)
	}

	return room, nil
}

func fetchRoomsBasedOnUserId(ctx context.Context, user_id string) ([]RoomLite, error) {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	roomsTable := pgx.Identifier{schema, "rooms"}.Sanitize()
	membersTable := pgx.Identifier{schema, "room_members"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT r.id, r.name, r.created_at, r.updated_at
		FROM %s r
		JOIN %s m ON r.id = m.room_id
		where m.user_id = $1
	`, roomsTable, membersTable)

	rows, err := Pool.Query(ctx, sql, user_id)
	if err != nil {
		return nil, fmt.Errorf("Query failed: %w", err)
	}
	defer rows.Close()

	var rooms []RoomLite

	for rows.Next() {
		var r RoomLite
		if err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.CreatedAt,
			&r.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("Failed to get room ID: %w", err)
		}
		rooms = append(rooms, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("There's an error during iterating rows array: %w", err)
	}

	return rooms, nil
}

func (app *App) fetchRoomMessage(ctx context.Context, room_id string, offset_id string) ([]RoomMessage, error) {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}
	messagesTable := pgx.Identifier{schema, "messages"}.Sanitize()
	usersTable := pgx.Identifier{schema, "users"}.Sanitize()
	var sql string
	var args []interface{}
	if offset_id == "" {
		sql = fmt.Sprintf(`
            SELECT m.id, COALESCE(u.username, 'Unknown User'), m.room_id, m.content, m.created_at
            FROM %s m
            LEFT JOIN %s u ON m.user_id = u.id
            WHERE m.room_id = $1
            ORDER BY m.id DESC
            LIMIT 50
        `, messagesTable, usersTable)
		args = []interface{}{room_id}
	} else {
		sql = fmt.Sprintf(`
            SELECT m.id, COALESCE(u.username, 'Unknown User'), m.room_id, m.content, m.created_at
            FROM %s m
            LEFT JOIN %s u ON m.user_id = u.id
            WHERE m.room_id = $1 AND m.id < $2
            ORDER BY m.id DESC
            LIMIT 50
        `, messagesTable, usersTable)
		args = []interface{}{room_id, offset_id}
	}

	rows, err := Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("Query failed: %w", err)
	}
	defer rows.Close()

	var messages []RoomMessage

	for rows.Next() {
		var m RoomMessage
		var createdAt time.Time
		if err := rows.Scan(
			&m.Id,
			&m.Owner_name,
			&m.Room_ID,
			&m.Content,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("Failed to get room message: %w", err)
		}
		m.TimeStamp = createdAt.Format(time.RFC3339)
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("There's an error during iterating rows array: %w", err)
	}

	return messages, nil
}

func addUserToRoom(ctx context.Context, userId string, invite_code string) error {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}

	//extract the room id
	room_invites_table := pgx.Identifier{schema, "room_invitations"}.Sanitize()
	room_invite_sql := fmt.Sprintf(`
		SELECT room_id FROM %s WHERE invite_code = $1
	`, room_invites_table)

	var room_id string
	err := Pool.QueryRow(ctx, room_invite_sql, invite_code).Scan(&room_id)

	if err != nil {
		return fmt.Errorf("Failed to query room invite %v", err)
	}

	if room_id == "" {
		return fmt.Errorf("Invite code invalid or room not found")
	}

	//make sure that the room exists
	var room_name string
	rooms_table := pgx.Identifier{schema, "rooms"}.Sanitize()
	rooms_table_sql := fmt.Sprintf(`
		SELECT name FROM %s WHERE id = $1
	`, rooms_table)

	err = Pool.QueryRow(ctx, rooms_table_sql, room_id).Scan(&room_name)

	if err != nil {
		return fmt.Errorf("Failed to query room %v", err)
	}

	if room_id == "" {
		return fmt.Errorf("Room not found")
	}

	// log.Println(room_id)
	// log.Println(room_name)

	//add user to room
	room_member_table := pgx.Identifier{schema, "room_members"}.Sanitize()
	sql := fmt.Sprintf(`
		INSERT INTO %s (user_id, room_id, joined_at)
		VALUES (@user_id, @room_id, @joined_at)
	`, room_member_table)

	_, err = Pool.Exec(ctx, sql, pgx.NamedArgs{
		"user_id":   userId,
		"room_id":   room_id,
		"joined_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("Failed to add user to room %v", err)
	}

	return nil
}

func addUserToRoomByID(ctx context.Context, userId string, room_id string) error {
	schema := "chat"
	if schema == "" {
		log.Println("Warning: DB_SCHEMA is not set, defaulting to 'public'")
		schema = "public"
	}

	//make sure that the room exists
	var room_name string
	rooms_table := pgx.Identifier{schema, "rooms"}.Sanitize()
	rooms_table_sql := fmt.Sprintf(`
		SELECT name FROM %s WHERE id = $1
	`, rooms_table)

	err := Pool.QueryRow(ctx, rooms_table_sql, room_id).Scan(&room_name)

	if err != nil {
		return fmt.Errorf("Failed to query room %v", err)
	}

	//add user to room
	room_member_table := pgx.Identifier{schema, "room_members"}.Sanitize()
	sql := fmt.Sprintf(`
		INSERT INTO %s (user_id, room_id, joined_at)
		VALUES (@user_id, @room_id, @joined_at)
	`, room_member_table)

	_, err = Pool.Exec(ctx, sql, pgx.NamedArgs{
		"user_id":   userId,
		"room_id":   room_id,
		"joined_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("Failed to add user to room %v", err)
	}

	return nil
}

func createNewRoom(ctx context.Context, userId string, room_name string) error {
	schema := "chat"
	var room_id string

	//create new room
	room_table := pgx.Identifier{schema, "rooms"}.Sanitize()
	sql := fmt.Sprintf(`
		INSERT INTO %s (name, owner_id)
		VALUES (@name, @owner_id)
		RETURNING id
	`, room_table)

	err := Pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"name":     room_name,
		"owner_id": userId,
	}).Scan(&room_id)

	if err != nil {
		return fmt.Errorf("Failed to query room %v", err)
	}

	addUserToRoomByID(ctx, userId, room_id)

	//If fail then generate code again when user need it
	//create new invite code
	MAX_TRIES := 10 // try 10 times

	for i := 0; i < MAX_TRIES; i++ {
		invite_code, err := createInviteCode()
		if err != nil {
			log.Printf("Failed to create invite code %v", err)
		}

		room_invites_table := pgx.Identifier{schema, "room_invitations"}.Sanitize()
		sql = fmt.Sprintf(`
		INSERT INTO %s (room_id, invited_by, invite_code)
		VALUES (@room_id, @invited_by, @invite_code)
		`, room_invites_table)

		_, err = Pool.Exec(ctx, sql, pgx.NamedArgs{
			"room_id":     room_id,
			"invited_by":  userId,
			"invite_code": invite_code,
		})

		//successfully saved
		if err == nil {
			return nil
		}

		log.Println("Duplicated key, trying again")
	}

	return nil
}
