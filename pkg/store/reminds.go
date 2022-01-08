package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Remind represents a Remind object.
type Remind struct {
	ID             primitive.ObjectID `bson:"_id"`
	DiscordUserID  string             `bson:"discordUserId"`
	PetName        string             `bson:"petName"`
	Character      string             `bson:"character"`
	MissedReminder int                `bson:"missedReminder"`
	NextRemind     time.Time          `bson:"nextRemind"`
	ReminderSent   bool               `bson:"reminderSent"`
	TimeoutRemind  time.Time          `bson:"timeoutRemind"`
}

// CreateRemind creates a new remind.
func (s *Store) CreateRemind(ctx context.Context, remind Remind) error {
	if _, err := s.reminds.InsertOne(ctx, remind); err != nil {
		return fmt.Errorf("create remind: %w", err)
	}

	return nil
}

// GetRemind gets a remind with the given ID.
func (s *Store) GetRemind(ctx context.Context, id string) (Remind, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Remind{}, fmt.Errorf("object id: %w", err)
	}

	var remind Remind
	if err = s.reminds.FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(&remind); err != nil {
		return Remind{}, fmt.Errorf("find remind: %w", err)
	}

	return remind, nil
}

// UpdateRemind updates the given remind.
func (s *Store) UpdateRemind(ctx context.Context, remind Remind) error {
	if _, err := s.reminds.UpdateOne(ctx, bson.D{{Key: "_id", Value: remind.ID}}, bson.D{{Key: "$set", Value: remind}}); err != nil {
		return fmt.Errorf("create remind: %w", err)
	}

	return nil
}

// ListReminds lists all the reminds.
func (s *Store) ListReminds(ctx context.Context) ([]Remind, error) {
	res, err := s.reminds.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find reminds: %w", err)
	}

	var reminds []Remind
	if err = res.All(ctx, &reminds); err != nil {
		return nil, fmt.Errorf("decode reminds: %w", err)
	}

	return reminds, nil
}

// RemoveRemind removes the remind with the given id.
func (s *Store) RemoveRemind(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("object id: %w", err)
	}

	if _, err = s.reminds.DeleteOne(ctx, bson.D{{Key: "_id", Value: objectID}}); err != nil {
		return fmt.Errorf("delete remind: %w", err)
	}

	return nil
}
