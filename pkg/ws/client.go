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
	CheckOrigin: func(r *http.Request) bool { return true }, // prod’da domain bilan cheklang
}

type Deps struct {
	Hub   *Hub
	Auth  *auth.Client
	Repo  *db.Repo
}

func ServeWS(d Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		token := q.Get("token")
		userIDStr := q.Get("user_id")
		courseIDStr := q.Get("course_id")
		studentIDStr := q.Get("student_id") // mentor bo‘lsa kerak bo‘ladi

		if token == "" || userIDStr == "" || courseIDStr == "" {
			http.Error(w, "missing query params", http.StatusBadRequest)
			return
		}
		userID, _ := strconv.ParseInt(userIDStr, 10, 64)
		courseID, _ := strconv.ParseInt(courseIDStr, 10, 64)

		u, err := d.Auth.GetUser(userID, token)
		if err != nil {
			http.Error(w, "auth failed", http.StatusUnauthorized)
			return
		}

		var mentorID, studentID int64
		if u.IsMentor {
			if studentIDStr == "" {
				http.Error(w, "student_id required for mentor", http.StatusBadRequest)
				return
			}
			studentID, _ = strconv.ParseInt(studentIDStr, 10, 64)
			mentorID = userID
		} else {
			mentorID, err = d.Auth.GetCourse(courseID, token)
			if err != nil || mentorID == 0 {
				http.Error(w, "cannot resolve mentor", http.StatusForbidden)
				return
			}
			studentID = userID
		}

		ctx := r.Context()
		conv, err := d.Repo.GetOrCreateConversation(ctx, courseID, mentorID, studentID)
		if err != nil {
			http.Error(w, "conversation error", 500)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil { return }
		defer ws.Close()

		c := &Client{
			conn:   ws,
			send:   make(chan []byte, 16),
			userID: userID,
			convID: conv.ID,
		}

		// write pump
		go func() {
			for msg := range c.send {
				c.conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
				_ = c.conn.WriteMessage(websocket.TextMessage, msg)
			}
		}()

		room := conv.ID.Hex()
		d.Hub.Join(room, c)
		defer func() { d.Hub.Leave(room, c); close(c.send) }()

		// read pump
		for {
			c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			_, raw, err := c.conn.ReadMessage()
			if err != nil { return }

			var in Inbound
			if err := json.Unmarshal(raw, &in); err != nil { continue }
			if in.Type != "message" || in.Text == "" { continue }

			msg, err := d.Repo.CreateMessage(context.Background(), c.convID, c.userID, in.Text)
			if err != nil { continue }

			out := Outbound{
				Type: "message",
				ConversationID: room,
				SenderID: c.userID,
				Text: msg.Text,
				CreatedAt: msg.CreatedAt.UTC().Format(time.RFC3339),
			}
			payload, _ := json.Marshal(out)

			// senderga ham, peerga ham
			c.send <- payload
			d.Hub.Broadcast(room, payload, c)
		}
	}
}
