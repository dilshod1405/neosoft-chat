package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	DB *mongo.Database
	convColl *mongo.Collection
	msgColl  *mongo.Collection
}

type Conversation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LessonID  int64              `json:"lesson_id"`
	MentorID  int64              `json:"mentor_id"`
	StudentID int64              `json:"student_id"`
	CreatedAt time.Time          `json:"created_at"`
}

type Message struct {
	ID             primitive.ObjectID `json:"id"`
	ConversationID primitive.ObjectID `json:"conversation_id"`
	SenderID       int64              `json:"sender_id"`
	Text           string             `json:"text"`
	CreatedAt      time.Time          `json:"created_at"`
}

func New(db *mongo.Database) *Repo {
	return &Repo{
		DB: db,
		convColl: db.Collection("conversations"),
		msgColl:  db.Collection("messages"),
	}
}

func (r *Repo) EnsureIndexes(ctx context.Context) error {
	_, err := r.convColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{"lesson_id", 1},
			{"mentor_id", 1},
			{"student_id", 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *Repo) GetOrCreateConversation(ctx context.Context, lessonID, mentorID, studentID int64) (*Conversation, error) {
	var conv Conversation
	err := r.convColl.FindOne(ctx, bson.M{
		"lesson_id":  lessonID,
		"mentor_id":  mentorID,
		"student_id": studentID,
	}).Decode(&conv)

	if err == nil {
		return &conv, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	conv = Conversation{
		LessonID: lessonID,
		MentorID: mentorID,
		StudentID: studentID,
		CreatedAt: time.Now(),
	}

	res, err := r.convColl.InsertOne(ctx, conv)
	if err != nil {
		return nil, err
	}

	conv.ID = res.InsertedID.(primitive.ObjectID)
	return &conv, nil
}

func (r *Repo) CreateMessage(ctx context.Context, convID primitive.ObjectID, sender int64, text string) (*Message, error) {
	msg := &Message{
		ConversationID: convID,
		SenderID: sender,
		Text: text,
		CreatedAt: time.Now(),
	}

	res, err := r.msgColl.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}

	msg.ID = res.InsertedID.(primitive.ObjectID)
	return msg, nil
}
