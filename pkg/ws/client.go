package ws

import (
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	userID int64
	convID primitive.ObjectID
	hub    *Hub
}
