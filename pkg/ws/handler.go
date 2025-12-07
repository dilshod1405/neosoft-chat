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

		user, err := d.Auth.GetUser(userID, token)
		if err != nil {
			http.Error(w, "auth failed", 401)
			return
		}

		lesson, err := d.Auth.GetLesson(lessonID, token)
		if err != nil {
			http.Error(w, "lesson not found", 404)
			return
		}

		var mentorID int64 = lesson.TeacherID
		var studentID int64

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

		client := &Client{
			conn:   wsconn,
			send:   make(chan []byte, 16),
			userID: userID,
			convID: conv.ID,
			hub:    d.Hub,
		}

		go func() {
			for msg := range client.send {
				client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				err := client.conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					return
				}
			}
		}()

		room := conv.ID.Hex()
		d.Hub.Join(room, client)

		defer func() {
			d.Hub.Leave(room, client)
			close(client.send)
			client.conn.Close()
		}()

		for {
			_, raw, err := client.conn.ReadMessage()
			if err != nil {
				return
			}

			var in Inbound
			if json.Unmarshal(raw, &in) != nil {
				continue
			}

			switch in.Type {

			case "typing":
				out := Outbound{
					Type:           "typing",
					ConversationID: room,
					SenderID:       client.userID,
					IsTyping:       in.IsTyping,
				}
				data, _ := json.Marshal(out)
				d.Hub.Broadcast(room, data, client)
				continue

			case "viewed":
				if in.MessageID == "" {
					continue
				}
				d.Repo.MarkMessageViewed(in.MessageID, client.userID)

				out := Outbound{
					Type:     "viewed",
					MessageID: in.MessageID,
					ViewerID:  client.userID,
				}
				data, _ := json.Marshal(out)
				d.Hub.Broadcast(room, data, client)
				continue
			}

			if in.Type == "message" && in.Text != "" {
				msg, err := d.Repo.CreateMessage(r.Context(), conv.ID, userID, in.Text)
				if err != nil {
					continue
				}

				out := Outbound{
					Type:           "message",
					ConversationID: room,
					SenderID:       userID,
					Text:           msg.Text,
					CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
				}

				data, _ := json.Marshal(out)
				client.send <- data
				d.Hub.Broadcast(room, data, client)
			}
		}
	}
}
