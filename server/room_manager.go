package main

import (
	"context"
	"fmt"
	"log"
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
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
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
			&r.Description,
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
