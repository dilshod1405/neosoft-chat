package http

import (
	"encoding/json"
	"net/http"
	"chat-service/pkg/db"
)

// GetMessages godoc
// @Summary Get chat messages
// @Description Returns all messages belonging to a conversation
// @Tags Messages
// @Produce json
// @Param conversation_id query string true "Conversation ID"
// @Success 200 {array} db.Message
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /messages [get]
func GetMessages(repo *db.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		convID := r.URL.Query().Get("conversation_id")
		if convID == "" {
			http.Error(w, "conversation_id required", http.StatusBadRequest)
			return
		}

		msgs, err := repo.GetMessages(convID)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(msgs)
	}
}
