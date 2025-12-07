package http

import (
	"encoding/json"
	"net/http"

	"chat-service/pkg/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetConversation godoc
// @Summary Get conversation detail
// @Tags Conversations
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} db.Conversation
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /conversations/{id} [get]
func GetConversation(repo *db.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		conv, err := repo.GetConversationByID(r.Context(), oid)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(conv)
	}
}
