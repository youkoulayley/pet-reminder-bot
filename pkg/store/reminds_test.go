package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStore_CreateRemind(t *testing.T) {
	ctx := context.Background()
	s := createStore(t, nil)

	remind := Remind{
		ID:             primitive.NewObjectID(),
		DiscordUserID:  "discordUser",
		PetName:        "pet",
		Character:      "character",
		MissedReminder: 0,
		NextRemind:     time.Time{},
		ReminderSent:   false,
		TimeoutRemind:  time.Time{},
	}
	err := s.CreateRemind(ctx, remind)
	require.NoError(t, err)

	var got Remind
	err = s.reminds.FindOne(ctx, bson.D{{Key: "_id", Value: remind.ID}}).Decode(&got)
	require.NoError(t, err)

	assert.Equal(t, remind, got)
}

func TestStore_GetRemind(t *testing.T) {
	ctx := context.Background()

	reminds := []Remind{
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser",
			PetName:        "pet",
			Character:      "character",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser2",
			PetName:        "pet2",
			Character:      "character2",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
	}
	s := createStore(t, reminds)

	remind, err := s.GetRemind(ctx, reminds[0].ID.Hex())
	require.NoError(t, err)

	assert.Equal(t, reminds[0], remind)
}

func TestStore_UpdateRemind(t *testing.T) {
	ctx := context.Background()

	reminds := []Remind{
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser",
			PetName:        "pet",
			Character:      "character",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser2",
			PetName:        "pet2",
			Character:      "character2",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
	}
	s := createStore(t, reminds)

	update := Remind{
		ID:             reminds[0].ID,
		DiscordUserID:  "123",
		PetName:        "petpet",
		Character:      "Toto",
		MissedReminder: 10,
		NextRemind:     time.Time{}.Add(time.Hour),
		ReminderSent:   true,
		TimeoutRemind:  time.Time{}.Add(2 * time.Hour),
	}

	err := s.UpdateRemind(ctx, update)
	require.NoError(t, err)

	var got Remind
	err = s.reminds.FindOne(ctx, bson.D{{Key: "_id", Value: reminds[0].ID}}).Decode(&got)
	require.NoError(t, err)

	assert.Equal(t, update, got)
}

func TestStore_ListReminds(t *testing.T) {
	ctx := context.Background()

	reminds := []Remind{
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser",
			PetName:        "pet",
			Character:      "character",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser2",
			PetName:        "pet2",
			Character:      "character2",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
	}
	s := createStore(t, reminds)

	got, err := s.ListReminds(ctx)
	require.NoError(t, err)

	assert.Equal(t, reminds, got)
}

func TestStore_RemoveRemind(t *testing.T) {
	ctx := context.Background()

	reminds := []Remind{
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser",
			PetName:        "pet",
			Character:      "character",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
		{
			ID:             primitive.NewObjectID(),
			DiscordUserID:  "discordUser2",
			PetName:        "pet2",
			Character:      "character2",
			MissedReminder: 0,
			NextRemind:     time.Time{},
			ReminderSent:   false,
			TimeoutRemind:  time.Time{},
		},
	}
	s := createStore(t, reminds)

	err := s.RemoveRemind(ctx, reminds[1].ID.Hex())
	require.NoError(t, err)

	res, err := s.reminds.Find(ctx, bson.D{})
	require.NoError(t, err)

	var got []Remind
	err = res.All(ctx, &got)
	require.NoError(t, err)

	assert.Equal(t, []Remind{reminds[0]}, got)
}
