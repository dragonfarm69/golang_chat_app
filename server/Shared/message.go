package shared

type RoomMessage struct {
	Id           string `json:"id"`
	Owner_name   string `json:"owner_name"`
	Room_ID      string `json:"room_id"`
	Content      string `json:"content"`
	Message_Type string `json:"message_type"`
	TimeStamp    string `json:"timeStamp"`
}

type MessagePayload struct {
	Id           string `json:"id"`
	User_ID      string `json:"user_id"`
	UserName     string `json:"username"`
	Room_ID      string `json:"room_id"`
	Content      string `json:"content"`
	TimeStamp    string `json:"timeStamp"`
	Action       string `json:"action"`
	Message_Type string `json:"message_type"`
}

type ResponseMessagePayload struct {
	OriginalId   string `json:"original_id"`
	Id           string `json:"id"`
	User_ID      string `json:"user_id"`
	UserName     string `json:"username"`
	Room_ID      string `json:"room_id"`
	Content      string `json:"content"`
	TimeStamp    string `json:"timeStamp"`
	Action       string `json:"action"`
	Message_Type string `json:"message_type"`
}

type FileMetaData struct {
	FileName string `json:"file_name"`
	FileSize string `json:"file_size"`
	FileType string `json:"file_type"`
}
