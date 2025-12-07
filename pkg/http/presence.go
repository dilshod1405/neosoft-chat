package http

import (
	"encoding/json"
	"net/http"
	"chat-service/pkg/ws"
	"strconv"
)

// GetPresence godoc
// @Summary Get user online/offline status
// @Tags Presence
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} map[string]bool
// @Router /presence [get]
func GetPresence(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("user_id")
		id, _ := strconv.ParseInt(idStr, 10, 64)

		status := hub.IsOnline(id)

		json.NewEncoder(w).Encode(map[string]bool{
			"online": status,
		})
	}
}
