// @Summary WebSocket Chat Endpoint
// @Description Connect to real-time chat WebSocket
// @Tags WebSocket
// @Param token query string true "JWT Token"
// @Param user_id query int true "User ID"
// @Param lesson_id query int true "Lesson ID"
// @Param student_id query int false "Student ID (mentor only)"
// @Router /ws [get]

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"chat-service/pkg/auth"
	"chat-service/pkg/db"
	httpapi "chat-service/pkg/http"
	"chat-service/pkg/ws"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB")
	django := os.Getenv("DJANGO_BASE_URL")
	port := os.Getenv("PORT")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	repo := db.New(client.Database(dbName))
	repo.EnsureIndexes(context.Background())

	hub := ws.NewHub()
	authClient := auth.New(django)

	http.HandleFunc("/ws", ws.ServeWS(ws.Deps{
		Hub:  hub,
		Auth: authClient,
		Repo: repo,
	}))

	http.HandleFunc("/messages", httpapi.GetMessages(repo))
	http.HandleFunc("/conversations", httpapi.GetConversations(repo))
	http.HandleFunc("/conversations/", httpapi.GetConversation(repo))
	http.HandleFunc("/presence", httpapi.GetPresence(hub))


	log.Println("Chat service running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
