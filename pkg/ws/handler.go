package ws

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"chat-service/pkg/auth"
	"chat-service/pkg/db"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Deps struct {
	Hub  *Hub
	Auth *auth.Client
	Repo *db.Repo
}

func ServeWS(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		token := q.Get("token")
		uidStr := q.Get("user_id")
		lessonStr := q.Get("lesson_id")
		studentStr := q.Get("student_id")

		if token == "" || uidStr == "" || lessonStr == "" {
			http.Error(w, "missing params", 400)
			return
		}

		userID, _ := strconv.ParseInt(uidStr, 10, 64)
		lessonID, _ := strconv.ParseInt(lessonStr, 10, 64)

		// ------- Check token -------
		user, err := d.Auth.GetUser(userID, token)
		if err != nil {
			http.Error(w, "auth failed", 401)
			return
		}

		// ------- Get lesson -------
		lesson, err := d.Auth.GetLesson(lessonID, token)
		if err != nil {
			http.Error(w, "lesson not found", 404)
			return
		}

		var mentorID int64 = lesson.TeacherID
		var studentID int64

		// ------- Mentor vs Student logic -------
		if user.IsMentor {
			if studentStr == "" {
				http.Error(w, "student_id required", 400)
				return
			}
			studentID, _ = strconv.ParseInt(studentStr, 10, 64)

			if mentorID != userID {
				http.Error(w, "not lesson mentor", 403)
				return
			}

		} else {
			studentID = userID
		}

		// ------- Find/create conversation -------
		conv, err := d.Repo.GetOrCreateConversation(r.Context(), lessonID, mentorID, studentID)
		if err != nil {
			http.Error(w, "conversation error", 500)
			return
		}

		wsconn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "ws upgrade failed", 500)
			return
		}
		defer wsconn.Close()

		client := &Client{
			conn: wsconn,
			send: make(chan []byte, 16),
			userID: userID,
			convID: conv.ID,
		}

		// writer goroutine
		go func() {
			for msg := range client.send {
				client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				_ = client.conn.WriteMessage(websocket.TextMessage, msg)
			}
		}()

		room := conv.ID.Hex()
		d.Hub.Join(room, client)
		defer func() {
			d.Hub.Leave(room, client)
			close(client.send)
		}()

		// reader loop
		for {
			_, raw, err := client.conn.ReadMessage()
			if err != nil {
				return
			}

			var in Inbound
			if json.Unmarshal(raw, &in) != nil {
				continue
			}
			if in.Type != "message" || in.Text == "" {
				continue
			}

			msg, err := d.Repo.CreateMessage(r.Context(), conv.ID, userID, in.Text)
			if err != nil {
				continue
			}

			out := Outbound{
				Type: "message",
				ConversationID: room,
				SenderID: userID,
				Text: msg.Text,
				CreatedAt: msg.CreatedAt.Format(time.RFC3339),
			}

			payload, _ := json.Marshal(out)

			client.send <- payload
			d.Hub.Broadcast(room, payload, client)
		}
	}
}
