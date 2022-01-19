package handlers

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/skwair/harmony"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestHandler_ReactionAdd(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	pet := store.Pet{
		Name:            "Chacha",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	d := &discordMock{}
	d.On("Message", "123").Return(&harmony.Message{Content: "ID: " + testRemindID}, nil).Once()

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{
		ID:             objectID,
		DiscordUserID:  testDiscordUserID,
		PetName:        "Chacha",
		Character:      "Test",
		MissedReminder: 0,
		NextRemind:     time.Time{},
		ReminderSent:   false,
		TimeoutRemind:  time.Time{},
	}, nil).Once()

	s.On("GetPet", "Chacha").Return(pet, nil).Once()

	s.On("UpdateRemind", mock.MatchedBy(func(remind store.Remind) bool {
		if time.Now().Add(pet.FoodMinDuration).Sub(remind.NextRemind) > time.Minute {
			return false
		}

		if time.Now().Add(pet.FoodMaxDuration).Sub(remind.TimeoutRemind) > time.Minute {
			return false
		}

		return reflect.DeepEqual(store.Remind{
			ID:             objectID,
			DiscordUserID:  testDiscordUserID,
			PetName:        "Chacha",
			Character:      "Test",
			MissedReminder: 0,
			NextRemind:     remind.NextRemind,
			ReminderSent:   false,
			TimeoutRemind:  remind.TimeoutRemind,
		}, remind)
	})).Return(nil).Once()

	r := &reminderMock{}
	r.On("SetUpdate").Once()

	h := Handler{store: s, reminder: r, discord: d}
	h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
}

func TestHandler_ReactionAdd_messageError(t *testing.T) {
	d := &discordMock{}
	d.On("Message", "123").Return(&harmony.Message{}, errors.New("boom")).Once()

	h := Handler{discord: d}
	h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
}

func TestHandler_ReactionAdd_validation(t *testing.T) {
	tests := []struct {
		desc    string
		message string
	}{
		{
			desc:    "invalid id",
			message: "ID: 123",
		},
		{
			desc: "invalid id",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On("Message", "123").Return(&harmony.Message{Content: test.message}, nil).Once()

			h := Handler{discord: d}
			h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
		})
	}
}

func TestHandler_ReactionAdd_getRemindError(t *testing.T) {
	d := &discordMock{}
	d.On("Message", "123").Return(&harmony.Message{Content: "ID: " + testRemindID}, nil).Once()

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{}, errors.New("boom")).Once()

	h := Handler{store: s, discord: d}
	h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
}

func TestHandler_ReactionAdd_badUser(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	d := &discordMock{}
	d.On("Message", "123").Return(&harmony.Message{Content: "ID: " + testRemindID}, nil).Once()

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{
		ID:             objectID,
		DiscordUserID:  "5",
		PetName:        "Chacha",
		Character:      "Test",
		MissedReminder: 0,
		NextRemind:     time.Time{},
		ReminderSent:   false,
		TimeoutRemind:  time.Time{},
	}, nil).Once()

	h := Handler{store: s, discord: d}
	h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
}

func TestHandler_ReactionAdd_getPetError(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	d := &discordMock{}
	d.On("Message", "123").Return(&harmony.Message{Content: "ID: " + testRemindID}, nil).Once()

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{
		ID:             objectID,
		DiscordUserID:  testDiscordUserID,
		PetName:        "Chacha",
		Character:      "Test",
		MissedReminder: 0,
		NextRemind:     time.Time{},
		ReminderSent:   false,
		TimeoutRemind:  time.Time{},
	}, nil).Once()

	s.On("GetPet", "Chacha").Return(store.Pet{}, errors.New("boom")).Once()

	h := Handler{store: s, discord: d}
	h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
}

func TestHandler_ReactionAdd_updateRemindError(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	pet := store.Pet{
		Name:            "Chacha",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	d := &discordMock{}
	d.On("Message", "123").Return(&harmony.Message{Content: "ID: " + testRemindID}, nil).Once()

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{
		ID:             objectID,
		DiscordUserID:  testDiscordUserID,
		PetName:        "Chacha",
		Character:      "Test",
		MissedReminder: 0,
		NextRemind:     time.Time{},
		ReminderSent:   false,
		TimeoutRemind:  time.Time{},
	}, nil).Once()

	s.On("GetPet", "Chacha").Return(pet, nil).Once()

	s.On("UpdateRemind", mock.MatchedBy(func(remind store.Remind) bool {
		if time.Now().Add(pet.FoodMinDuration).Sub(remind.NextRemind) > time.Minute {
			return false
		}

		if time.Now().Add(pet.FoodMaxDuration).Sub(remind.TimeoutRemind) > time.Minute {
			return false
		}

		return reflect.DeepEqual(store.Remind{
			ID:             objectID,
			DiscordUserID:  testDiscordUserID,
			PetName:        "Chacha",
			Character:      "Test",
			MissedReminder: 0,
			NextRemind:     remind.NextRemind,
			ReminderSent:   false,
			TimeoutRemind:  remind.TimeoutRemind,
		}, remind)
	})).Return(errors.New("boom")).Once()

	h := Handler{store: s, discord: d}
	h.ReactionAdd(&harmony.MessageReaction{MessageID: "123", UserID: testDiscordUserID})
}
