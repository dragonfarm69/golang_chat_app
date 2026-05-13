package data

import (
	shared "chat-app-server/Shared"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
)

const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func CreateInviteCode() (string, error) {
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

func (db *DataStorage) FetchFullRoomDataBasedOnRoomId(ctx context.Context, room_id string) (shared.Room, error) {
	table := pgx.Identifier{db.schema, "rooms"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT * FROM %s WHERE id = $1
	`, table)

	var room shared.Room
	err := db.Db_pool.QueryRow(ctx, sql, room_id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.IsPrivate,
		&room.OwnerID,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return shared.Room{}, fmt.Errorf("failed to fetch room_id: %w", err)
	}

	return room, nil
}

func (db *DataStorage) FetchRoomLiteDataBasedOnRoomId(ctx context.Context, room_id string) (shared.RoomLite, error) {
	table := pgx.Identifier{db.schema, "rooms"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT * FROM %s WHERE id = $1
	`, table)

	var room shared.RoomLite
	err := db.Db_pool.QueryRow(ctx, sql, room_id).Scan(
		&room.ID,
		&room.Name,
		&room.Description,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return shared.RoomLite{}, fmt.Errorf("failed to fetch room_id: %w", err)
	}

	return room, nil
}

func (db *DataStorage) FetchRoomsBasedOnUserId(ctx context.Context, user_id string) ([]shared.RoomLite, error) {
	roomsTable := pgx.Identifier{db.schema, "rooms"}.Sanitize()
	membersTable := pgx.Identifier{db.schema, "room_members"}.Sanitize()

	sql := fmt.Sprintf(`
		SELECT r.id, r.name, r.created_at, r.updated_at
		FROM %s r
		JOIN %s m ON r.id = m.room_id
		where m.user_id = $1
	`, roomsTable, membersTable)

	rows, err := db.Db_pool.Query(ctx, sql, user_id)
	if err != nil {
		return nil, fmt.Errorf("Query failed: %w", err)
	}
	defer rows.Close()

	var rooms []shared.RoomLite

	for rows.Next() {
		var r shared.RoomLite
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

func (db *DataStorage) FetchRoomMessage(ctx context.Context, room_id string, offset_id string) ([]shared.RoomMessage, error) {
	log.Println("FETCHING MESSAGE WITH OFFSET: ", offset_id)
	//check in redis first
	// key := fmt.Sprintf("room:%s:recent_messages", room_id)
	// redis_data, err := app.redis_db.LRange(ctx, key, 0, 49).Result()

	// if offset_id == "" && err == nil && len(redis_data) > 0 {
	// 	log.Println("Getting data from redis for room: ", room_id)
	// 	var messages []RoomMessage
	// 	for _, item := range redis_data {
	// 		var m RoomMessage
	// 		if err := json.Unmarshal([]byte(item), &m); err == nil {
	// 			messages = append(messages, m)
	// 		} else {
	// 			log.Println("Failed to unmarshal cached data")
	// 		}
	// 	}
	// 	return messages, nil
	// }

	messagesTable := pgx.Identifier{db.schema, "messages"}.Sanitize()
	usersTable := pgx.Identifier{db.schema, "users"}.Sanitize()
	var sql string
	var args []interface{}
	if offset_id == "" {
		sql = fmt.Sprintf(`
            SELECT m.id, COALESCE(u.username, 'Unknown User'), m.room_id, m.content, m.created_at, m.message_type
            FROM %s m
            LEFT JOIN %s u ON m.user_id = u.id
            WHERE m.room_id = $1
            ORDER BY m.id DESC
            LIMIT 50
        `, messagesTable, usersTable)
		args = []interface{}{room_id}
	} else {
		sql = fmt.Sprintf(`
            SELECT m.id, COALESCE(u.username, 'Unknown User'), m.room_id, m.content, m.created_at, m.message_type
            FROM %s m
            LEFT JOIN %s u ON m.user_id = u.id
            WHERE m.room_id = $1 AND m.id < $2
            ORDER BY m.id DESC
            LIMIT 50
        `, messagesTable, usersTable)
		args = []interface{}{room_id, offset_id}
	}

	rows, err := db.Db_pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("Query failed: %w", err)
	}
	defer rows.Close()

	var messages []shared.RoomMessage

	for rows.Next() {
		var m shared.RoomMessage
		var createdAt time.Time
		if err := rows.Scan(
			&m.Id,
			&m.Owner_name,
			&m.Room_ID,
			&m.Content,
			&createdAt,
			&m.Message_Type,
		); err != nil {
			return nil, fmt.Errorf("Failed to get room message: %w", err)
		}
		m.TimeStamp = createdAt.Format(time.RFC3339)
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("There's an error during iterating rows array: %w", err)
	}

	log.Println("MESSAGES: ", messages)

	// //only cache if offset_id == "" -> need newest message
	// if offset_id == "" && len(messages) > 0 {
	// 	app.redis_db.Del(ctx, key)

	// 	var redis_cache_interface []interface{}
	// 	for _, m := range messages {
	// 		msg, _ := json.Marshal(m)
	// 		redis_cache_interface = append(redis_cache_interface, msg)
	// 	}
	// 	app.redis_db.RPush(ctx, key, redis_cache_interface...)
	// 	//keep only 50 newest message (could be bigger but that will come later)
	// 	app.redis_db.LTrim(ctx, key, 0, 49)
	// }

	return messages, nil
}

func (db *DataStorage) AddUserToRoom(ctx context.Context, userId string, invite_code string) error {
	//extract the room id
	room_invites_table := pgx.Identifier{db.schema, "room_invitations"}.Sanitize()
	room_invite_sql := fmt.Sprintf(`
		SELECT room_id FROM %s WHERE invite_code = $1
	`, room_invites_table)

	var room_id string
	err := db.Db_pool.QueryRow(ctx, room_invite_sql, invite_code).Scan(&room_id)

	if err != nil {
		return fmt.Errorf("Failed to query room invite %v", err)
	}

	if room_id == "" {
		return fmt.Errorf("Invite code invalid or room not found")
	}

	//make sure that the room exists
	var room_name string
	rooms_table := pgx.Identifier{db.schema, "rooms"}.Sanitize()
	rooms_table_sql := fmt.Sprintf(`
		SELECT name FROM %s WHERE id = $1
	`, rooms_table)

	err = db.Db_pool.QueryRow(ctx, rooms_table_sql, room_id).Scan(&room_name)

	if err != nil {
		return fmt.Errorf("Failed to query room %v", err)
	}

	if room_id == "" {
		return fmt.Errorf("Room not found")
	}

	// log.Println(room_id)
	// log.Println(room_name)

	//add user to room
	room_member_table := pgx.Identifier{db.schema, "room_members"}.Sanitize()
	sql := fmt.Sprintf(`
		INSERT INTO %s (user_id, room_id, joined_at)
		VALUES (@user_id, @room_id, @joined_at)
	`, room_member_table)

	_, err = db.Db_pool.Exec(ctx, sql, pgx.NamedArgs{
		"user_id":   userId,
		"room_id":   room_id,
		"joined_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("Failed to add user to room %v", err)
	}

	return nil
}

func (db *DataStorage) AddUserToRoomByID(ctx context.Context, userId string, room_id string) error {
	//make sure that the room exists
	var room_name string
	rooms_table := pgx.Identifier{db.schema, "rooms"}.Sanitize()
	rooms_table_sql := fmt.Sprintf(`
		SELECT name FROM %s WHERE id = $1
	`, rooms_table)

	err := db.Db_pool.QueryRow(ctx, rooms_table_sql, room_id).Scan(&room_name)

	if err != nil {
		return fmt.Errorf("Failed to query room %v", err)
	}

	//add user to room
	room_member_table := pgx.Identifier{db.schema, "room_members"}.Sanitize()
	sql := fmt.Sprintf(`
		INSERT INTO %s (user_id, room_id, joined_at)
		VALUES (@user_id, @room_id, @joined_at)
	`, room_member_table)

	_, err = db.Db_pool.Exec(ctx, sql, pgx.NamedArgs{
		"user_id":   userId,
		"room_id":   room_id,
		"joined_at": time.Now(),
	})

	if err != nil {
		return fmt.Errorf("Failed to add user to room %v", err)
	}

	return nil
}

func (db *DataStorage) CreateNewRoom(ctx context.Context, userId string, room_name string) error {
	var room_id string

	//create new room
	room_table := pgx.Identifier{db.schema, "rooms"}.Sanitize()
	sql := fmt.Sprintf(`
		INSERT INTO %s (name, owner_id)
		VALUES (@name, @owner_id)
		RETURNING id
	`, room_table)

	err := db.Db_pool.QueryRow(ctx, sql, pgx.NamedArgs{
		"name":     room_name,
		"owner_id": userId,
	}).Scan(&room_id)

	if err != nil {
		return fmt.Errorf("Failed to query room %v", err)
	}

	err = db.AddUserToRoomByID(ctx, userId, room_id)
	if err != nil {
		return fmt.Errorf("Failed to add user to room %v", err)
	}

	//If fail then generate code again when user need it
	//create new invite code
	MAX_TRIES := 10 // try 10 times

	for i := 0; i < MAX_TRIES; i++ {
		invite_code, err := CreateInviteCode()
		if err != nil {
			log.Printf("Failed to create invite code %v", err)
		}

		room_invites_table := pgx.Identifier{db.schema, "room_invitations"}.Sanitize()
		sql = fmt.Sprintf(`
		INSERT INTO %s (room_id, invited_by, invite_code)
		VALUES (@room_id, @invited_by, @invite_code)
		`, room_invites_table)

		_, err = db.Db_pool.Exec(ctx, sql, pgx.NamedArgs{
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
