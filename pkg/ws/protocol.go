package ws

type Inbound struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Outbound struct {
	Type string `json:"type"`
	ConversationID string `json:"conversation_id"`
	SenderID int64 `json:"sender_id"`
	Text string `json:"text"`
	CreatedAt string `json:"created_at"`
}
