package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat-service/pkg/auth"
	"chat-service/pkg/db"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	convID primitive.ObjectID
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // prod’da domain bilan cheklash kerak
}

type Deps struct {
	Hub  *Hub
	Auth *auth.Client
	Repo *db.Repo
}

// ✅ JSON error helper
func writeJSONError(w http.ResponseWriter, code, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"type":    "error",
		"code":    code,
		"message": msg,
	})
}

func ServeWS(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		token := q.Get("token")
		userIDStr := q.Get("user_id")
		courseIDStr := q.Get("course_id")
		studentIDStr := q.Get("student_id") // mentor bo‘lsa kerak bo‘ladi

		if token == "" || userIDStr == "" || courseIDStr == "" {
			writeJSONError(w, "missing_params", "token, user_id and course_id are required", 400)
			return
		}

		userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		courseID, _ := strconv.ParseInt(courseIDStr, 10, 64)

		// ✅ 1. Auth check
		u, err := d.Auth.GetUser(userID, token)
		if err != nil {
			writeJSONError(w, "auth_failed", "Invalid or expired token", 401)
			return
		}

		var mentorID, studentID int64

		// ✅ STUDENT LOGIC
		if !u.IsMentor {
			studentID = userID

			// ✅ Student shu kursda bormi?
			enrolled, err := d.Auth.CheckEnrollment(courseID, studentID, token)
			if err != nil || !enrolled {
				writeJSONError(w, "not_enrolled", "User is not enrolled in this course", 403)
				return
			}

			// ✅ Kurs mentori kim?
			mentorID, err = d.Auth.GetCourse(courseID, token)
			if err != nil || mentorID == 0 {
				writeJSONError(w, "cannot_resolve_mentor", "Cannot find mentor for this course", 403)
				return
			}

		} else {
			// ✅ MENTOR LOGIC
			if studentIDStr == "" {
				writeJSONError(w, "missing_params", "student_id required for mentor", 400)
				return
			}
			studentID, _ = strconv.ParseInt(studentIDStr, 10, 64)
			mentorID = userID

			// ✅ Kurs mentori aynan shu usermi?
			courseMentorID, err := d.Auth.GetCourse(courseID, token)
			if err != nil || courseMentorID != mentorID {
				writeJSONError(w, "not_course_mentor", "Mentor does not own this course", 403)
				return
			}

			// ✅ Student shu kursni sotib olganmi?
			enrolled, err := d.Auth.CheckEnrollment(courseID, studentID, token)
			if err != nil || !enrolled {
				writeJSONError(w, "student_not_enrolled", "Student is not part of this course", 403)
				return
			}
		}

		// ✅ 4. Conversation yaratish / olish
		ctx := r.Context()
		conv, err := d.Repo.GetOrCreateConversation(ctx, courseID, mentorID, studentID)
		if err != nil {
			writeJSONError(w, "conversation_error", "Failed to open conversation", 500)
			return
		}

		// ✅ 5. WebSocket upgrade
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			writeJSONError(w, "ws_upgrade_failed", "Failed to upgrade to WebSocket", 500)
			return
		}
		defer ws.Close()

		c := &Client{
			conn:   ws,
			send:   make(chan []byte, 16),
			userID: userID,
			convID: conv.ID,
		}

		// ✅ Writer goroutine
		go func() {
			for msg := range c.send {
				c.conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
				_ = c.conn.WriteMessage(websocket.TextMessage, msg)
			}
		}()

		room := conv.ID.Hex()
		d.Hub.Join(room, c)
		defer func() { d.Hub.Leave(room, c); close(c.send) }()

		// ✅ Read pump
		for {
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, raw, err := c.conn.ReadMessage()
			if err != nil {
				return
			}

			var in Inbound
			if err := json.Unmarshal(raw, &in); err != nil || in.Type != "message" || in.Text == "" {
				continue
			}

			msg, err := d.Repo.CreateMessage(context.Background(), c.convID, c.userID, in.Text)
			if err != nil {
				continue
			}

			out := Outbound{
				Type:          "message",
				ConversationID: room,
				SenderID:      c.userID,
				Text:          msg.Text,
				CreatedAt:     msg.CreatedAt.UTC().Format(time.RFC3339),
			}
			payload, _ := json.Marshal(out)

			c.send <- payload
			d.Hub.Broadcast(room, payload, c)
		}
	}
}
