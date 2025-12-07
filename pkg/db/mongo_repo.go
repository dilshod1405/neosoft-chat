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
	DB       *mongo.Database
	convColl *mongo.Collection
	msgColl  *mongo.Collection
}

type Conversation struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LessonID  int64              `bson:"lesson_id" json:"lesson_id"`
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
	ViewedBy       int64              `bson:"viewed_by,omitempty" json:"viewed_by,omitempty"`
}

func New(db *mongo.Database) *Repo {
	return &Repo{
		DB:       db,
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
		LessonID:  lessonID,
		MentorID:  mentorID,
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
		SenderID:       sender,
		Text:           text,
		CreatedAt:      time.Now(),
	}

	res, err := r.msgColl.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}

	msg.ID = res.InsertedID.(primitive.ObjectID)
	return msg, nil
}

func (r *Repo) GetMessages(convID string) ([]Message, error) {
	oid, _ := primitive.ObjectIDFromHex(convID)

	cur, err := r.msgColl.Find(context.Background(), bson.M{
		"conversation_id": oid,
	})
	if err != nil {
		return nil, err
	}

	var msgs []Message
	err = cur.All(context.Background(), &msgs)
	return msgs, err
}

func (r *Repo) MarkMessageViewed(msgID string, userID int64) error {
	oid, err := primitive.ObjectIDFromHex(msgID)
	if err != nil {
		return err
	}

	_, err = r.msgColl.UpdateByID(context.Background(), oid, bson.M{
		"$set": bson.M{"viewed_by": userID},
	})

	return err
}



func (r *Repo) ListConversations(filter bson.M) ([]Conversation, error) {
	cur, err := r.convColl.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var list []Conversation
	err = cur.All(context.Background(), &list)
	return list, err
}

func (r *Repo) GetConversationByID(ctx context.Context, id primitive.ObjectID) (*Conversation, error) {
	var conv Conversation
	err := r.convColl.FindOne(ctx, bson.M{"_id": id}).Decode(&conv)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}