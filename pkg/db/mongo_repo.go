package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	DB            *mongo.Database
	convColl      *mongo.Collection
	messageColl   *mongo.Collection
}

type Conversation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CourseID  int64              `bson:"course_id" json:"course_id"`
	MentorID  int64              `bson:"mentor_id" json:"mentor_id"`
	StudentID int64              `bson:"student_id" json:"student_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Message struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ConversationID primitive.ObjectID `bson:"conversation_id" json:"conversation_id"`
	SenderID       int64              `bson:"sender_id" json:"sender_id"`
	Text           string             `bson:"text" json:"text"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	IsRead         bool               `bson:"is_read" json:"is_read"`
}

func New(db *mongo.Database) *Repo {
	return &Repo{
		DB:          db,
		convColl:    db.Collection("conversations"),
		messageColl: db.Collection("messages"),
	}
}

// Call on startup once
func (r *Repo) EnsureIndexes(ctx context.Context) error {
	// conversations unique (course_id, mentor_id, student_id)
	_, err := r.convColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "course_id", Value: 1},
			{Key: "mentor_id", Value: 1},
			{Key: "student_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil { return err }

	// messages: query by conversation + created_at desc
	_, err = r.messageColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "conversation_id", Value: 1},
			{Key: "created_at", Value: -1},
		},
	})
	return err
}

func (r *Repo) GetOrCreateConversation(ctx context.Context, courseID, mentorID, studentID int64) (*Conversation, error) {
	var c Conversation
	err := r.convColl.FindOne(ctx, bson.M{
		"course_id":  courseID,
		"mentor_id":  mentorID,
		"student_id": studentID,
	}).Decode(&c)
	if err == nil {
		return &c, nil
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}
	c = Conversation{
		CourseID:  courseID,
		MentorID:  mentorID,
		StudentID: studentID,
		CreatedAt: time.Now().UTC(),
	}
	res, err := r.convColl.InsertOne(ctx, c)
	if err != nil { return nil, err }
	c.ID = res.InsertedID.(primitive.ObjectID)
	return &c, nil
}

func (r *Repo) CreateMessage(ctx context.Context, convID primitive.ObjectID, senderID int64, text string) (*Message, error) {
	m := &Message{
		ConversationID: convID,
		SenderID:       senderID,
		Text:           text,
		CreatedAt:      time.Now().UTC(),
		IsRead:         false,
	}
	res, err := r.messageColl.InsertOne(ctx, m)
	if err != nil { return nil, err }
	m.ID = res.InsertedID.(primitive.ObjectID)
	return m, nil
}

func (r *Repo) ListMessages(ctx context.Context, convID primitive.ObjectID, limit int64) ([]Message, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit)
	cur, err := r.messageColl.Find(ctx, bson.M{"conversation_id": convID}, opts)
	if err != nil { return nil, err }
	defer cur.Close(ctx)
	var out []Message
	for cur.Next(ctx) {
		var m Message
		if err := cur.Decode(&m); err != nil { return nil, err }
		out = append(out, m)
	}
	return out, cur.Err()
}
