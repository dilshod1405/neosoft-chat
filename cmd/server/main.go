package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"chat-service/pkg/auth"
	"chat-service/pkg/db"
	"chat-service/pkg/ws"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func env(k, def string) string { v := os.Getenv(k); if v == "" { return def }; return v }

func main() {
	port := env("PORT", "8080")
	mongoURI := env("MONGO_URI", "mongodb://localhost:27017")
	dbName := env("MONGO_DB", "chat")
	djangoBase := env("DJANGO_BASE_URL", "http://localhost:8000")

	// Mongo connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil { log.Fatal(err) }
	dbx := client.Database(dbName)

	repo := db.New(dbx)
	if err := repo.EnsureIndexes(context.Background()); err != nil { log.Fatal(err) }

	hub := ws.NewHub()
	authClient := auth.New(djangoBase)

	// WS endpoint
	http.HandleFunc("/ws", ws.ServeWS(ws.Deps{
		Hub:  hub,
		Auth: authClient,
		Repo: repo,
	}))

	// History endpoint (conversation_id=hex, limit=int)
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		convHex := q.Get("conversation_id")
		if convHex == "" { http.Error(w, "conversation_id required", 400); return }
		limitStr := q.Get("limit"); if limitStr == "" { limitStr = "50" }
		limit, _ := strconv.ParseInt(limitStr, 10, 64)

		oid, err := primitive.ObjectIDFromHex(convHex)
		if err != nil { http.Error(w, "bad conversation_id", 400); return }

		msgs, err := repo.ListMessages(r.Context(), oid, limit)
		if err != nil { http.Error(w, "db error", 500); return }

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msgs)
	})

	log.Println("chat server on :"+port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
