package shared

type WsEvent struct {
	Type    string `json:"action"`
	Room_ID string `json:"room_id"`
	Payload any    `json:"payload"`
}
