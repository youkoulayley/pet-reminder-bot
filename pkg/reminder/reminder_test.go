package reminder

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/skwair/harmony/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestReminder_SetUpdate(t *testing.T) {
	r, err := New(nil, nil)
	require.NoError(t, err)

	assert.Equal(t, true, r.needUpdate.Load())

	r.needUpdate.Store(false)
	assert.Equal(t, false, r.needUpdate.Load())

	r.SetUpdate()

	assert.Equal(t, true, r.needUpdate.Load())
}

func TestReminder_Process_loadRemindsError(t *testing.T) {
	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{}, errors.New("boom")).Once()

	r, err := New(s, nil)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
}

func TestReminder_Process_sendRemind(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		TimeoutRemind: time.Now().Add(time.Hour),
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Twice()

	d := &discordMock{}
	d.On("SendMessage", fmt.Sprintf("<@discordUser> Il faut nourrir \"pet\" sur character\nID: %s", id.Hex())).
		Return(&discord.Message{}, nil).
		Once()

	updatedRemind := remind
	updatedRemind.ReminderSent = true
	s.On("UpdateRemind", updatedRemind).Return(nil).Once()

	r, err := New(s, d)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestReminder_Process_sendRemind_sendMessageError(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		TimeoutRemind: time.Now().Add(time.Hour),
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Once()

	d := &discordMock{}
	d.On("SendMessage", fmt.Sprintf("<@discordUser> Il faut nourrir \"pet\" sur character\nID: %s", id.Hex())).
		Return(&discord.Message{}, errors.New("boom")).
		Once()

	r, err := New(s, d)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestReminder_Process_sendRemind_updateRemindError(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		TimeoutRemind: time.Now().Add(time.Hour),
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Once()

	d := &discordMock{}
	d.On("SendMessage", fmt.Sprintf("<@discordUser> Il faut nourrir \"pet\" sur character\nID: %s", id.Hex())).
		Return(&discord.Message{}, nil).
		Once()

	updatedRemind := remind
	updatedRemind.ReminderSent = true
	s.On("UpdateRemind", updatedRemind).Return(errors.New("boom")).Once()

	r, err := New(s, d)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestReminder_Process_sendTimeoutRemind(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		ReminderSent:  true,
	}

	pet := store.Pet{
		Name:            "pet",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Twice()
	s.On("GetPet", "pet").Return(pet, nil).Once()

	updatedRemind := remind
	updatedRemind.ReminderSent = false
	updatedRemind.MissedReminder = 1

	s.On("UpdateRemind", mock.MatchedBy(func(r store.Remind) bool {
		if time.Now().Add(pet.FoodMinDuration).Sub(r.NextRemind) > time.Minute {
			return false
		}

		if time.Now().Add(pet.FoodMaxDuration).Sub(r.TimeoutRemind) > time.Minute {
			return false
		}

		updatedRemind.NextRemind = r.NextRemind
		updatedRemind.TimeoutRemind = r.TimeoutRemind

		return reflect.DeepEqual(updatedRemind, r)
	})).Return(nil).Once()

	d := &discordMock{}
	d.On("SendMessage", mock.MatchedBy(func(msg string) bool {
		parts := strings.Split(msg, "\n")
		reminderPart := parts[0]
		datePart := parts[1]
		idPart := parts[2]

		if reminderPart != "<@discordUser> \"pet\" sur character a râté 1 repas." {
			return false
		}

		dateParts := strings.SplitN(datePart, ": ", 2)
		date, err := time.Parse(time.RFC1123, dateParts[1])
		if err != nil {
			return false
		}

		if time.Now().Add(pet.FoodMinDuration).Sub(date) > time.Minute {
			return false
		}

		return idPart == "ID: "+id.Hex()
	})).Return(&discord.Message{}, nil).
		Once()

	r, err := New(s, d)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
	d.AssertExpectations(t)
}

func TestReminder_Process_sendTimeoutRemind_getPetError(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		ReminderSent:  true,
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Once()
	s.On("GetPet", "pet").Return(store.Pet{}, errors.New("boom")).Once()

	r, err := New(s, nil)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
}

func TestReminder_Process_sendTimeoutRemind_updateRemindError(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		ReminderSent:  true,
	}

	pet := store.Pet{
		Name:            "pet",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Once()
	s.On("GetPet", "pet").Return(pet, nil).Once()

	updatedRemind := remind
	updatedRemind.ReminderSent = false
	updatedRemind.MissedReminder = 1

	s.On("UpdateRemind", mock.MatchedBy(func(r store.Remind) bool {
		if time.Now().Add(pet.FoodMinDuration).Sub(r.NextRemind) > time.Minute {
			return false
		}

		if time.Now().Add(pet.FoodMaxDuration).Sub(r.TimeoutRemind) > time.Minute {
			return false
		}

		updatedRemind.NextRemind = r.NextRemind
		updatedRemind.TimeoutRemind = r.TimeoutRemind

		return reflect.DeepEqual(updatedRemind, r)
	})).Return(errors.New("boom")).Once()

	r, err := New(s, nil)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
}

func TestReminder_Process_sendTimeoutRemind_sendMessageError(t *testing.T) {
	id := primitive.NewObjectID()

	remind := store.Remind{
		ID:            id,
		DiscordUserID: "discordUser",
		PetName:       "pet",
		Character:     "character",
		ReminderSent:  true,
	}

	pet := store.Pet{
		Name:            "pet",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	s := &storerMock{}
	s.On("ListReminds").Return([]store.Remind{remind}, nil).Twice()
	s.On("GetPet", "pet").Return(pet, nil).Once()

	updatedRemind := remind
	updatedRemind.ReminderSent = false
	updatedRemind.MissedReminder = 1

	s.On("UpdateRemind", mock.MatchedBy(func(r store.Remind) bool {
		if time.Now().Add(pet.FoodMinDuration).Sub(r.NextRemind) > time.Minute {
			return false
		}

		if time.Now().Add(pet.FoodMaxDuration).Sub(r.TimeoutRemind) > time.Minute {
			return false
		}

		updatedRemind.NextRemind = r.NextRemind
		updatedRemind.TimeoutRemind = r.TimeoutRemind

		return reflect.DeepEqual(updatedRemind, r)
	})).Return(nil).Once()

	d := &discordMock{}
	d.On("SendMessage", mock.MatchedBy(func(msg string) bool {
		parts := strings.Split(msg, "\n")
		reminderPart := parts[0]
		datePart := parts[1]
		idPart := parts[2]

		if reminderPart != "<@discordUser> \"pet\" sur character a râté 1 repas." {
			return false
		}

		dateParts := strings.SplitN(datePart, ": ", 2)
		date, err := time.Parse(time.RFC1123, dateParts[1])
		if err != nil {
			return false
		}

		if time.Now().Add(pet.FoodMinDuration).Sub(date) > time.Minute {
			return false
		}

		return idPart == "ID: "+id.Hex()
	})).Return(&discord.Message{}, errors.New("boom")).
		Once()

	r, err := New(s, d)
	require.NoError(t, err)

	r.Process(context.Background())

	s.AssertExpectations(t)
	d.AssertExpectations(t)
}
