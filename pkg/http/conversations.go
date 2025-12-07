package http

import (
	"encoding/json"
	"net/http"
	"chat-service/pkg/db"
	"strconv"
)

// GetConversations godoc
// @Summary Get list of conversations
// @Description Mentor → all students; Student → their own conversation
// @Tags Conversations
// @Produce json
// @Param mentor_id query int false "Mentor ID"
// @Param student_id query int false "Student ID"
// @Success 200 {array} db.Conversation
// @Failure 400 {object} map[string]string
// @Router /conversations [get]
func GetConversations(repo *db.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		mentorStr := r.URL.Query().Get("mentor_id")
		studentStr := r.URL.Query().Get("student_id")

		filter := map[string]interface{}{}

		if mentorStr != "" {
			mentorID, _ := strconv.ParseInt(mentorStr, 10, 64)
			filter["mentor_id"] = mentorID
		}

		if studentStr != "" {
			studentID, _ := strconv.ParseInt(studentStr, 10, 64)
			filter["student_id"] = studentID
		}

		list, err := repo.ListConversations(filter)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(list)
	}
}
