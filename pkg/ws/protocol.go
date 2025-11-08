package ws

type Inbound struct {
	Type string `json:"type"` // "message"
	Text string `json:"text,omitempty"`
}

type Outbound struct {
	Type           string `json:"type"` // "message"
	ConversationID string `json:"conversation_id,omitempty"`
	SenderID       int64  `json:"sender_id,omitempty"`
	Text           string `json:"text,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
}
