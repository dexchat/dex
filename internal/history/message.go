package history

type LogEntry struct {
	ReceivedAt int64   `json:"received_at"`
	ServerTime int64   `json:"server_time"`
	MsgID      *string `json:"msg_id"` // IRCv3 msgid, nil if the server does not support it
	Username   string  `json:"username"`
	Text       string  `json:"text"`
}
