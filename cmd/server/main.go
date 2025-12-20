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

	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	repo := db.New(client.Database(dbName))
	repo.EnsureIndexes(context.Background())

	hub := ws.NewHub()
	authClient := auth.New(django)

	// 🔌 WebSocket
	http.Handle("/ws", withCORS(ws.ServeWS(ws.Deps{
		Hub:  hub,
		Auth: authClient,
		Repo: repo,
	})))

	// 📦 REST API
	http.Handle("/messages", withCORS(httpapi.GetMessages(repo)))
	http.Handle("/conversations", withCORS(httpapi.GetConversations(repo)))
	http.Handle("/conversations/", withCORS(httpapi.GetConversation(repo)))
	http.Handle("/presence", withCORS(httpapi.GetPresence(hub)))

	log.Println("🚀 Chat service running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// 🌐 CORS middleware
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
