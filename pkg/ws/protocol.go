package ws

type Inbound struct {
    Type           string `json:"type"`
    Text           string `json:"text,omitempty"`
    MessageID      string `json:"message_id,omitempty"`
    ConversationID string `json:"conversation_id,omitempty"`
    IsTyping       bool   `json:"is_typing,omitempty"`
}

type Outbound struct {
    Type           string `json:"type"`
    ConversationID string `json:"conversation_id,omitempty"`
    SenderID       int64  `json:"sender_id,omitempty"`
    Text           string `json:"text,omitempty"`
    CreatedAt      string `json:"created_at,omitempty"`

    MessageID string `json:"message_id,omitempty"`
    ViewerID  int64  `json:"viewer_id,omitempty"`

    Online   bool `json:"online,omitempty"`
    IsTyping bool `json:"is_typing,omitempty"`
}
