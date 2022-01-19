package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/skwair/harmony"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	testDiscordUserID = "2"
	testRemindID      = "61e71f03735c4de773d8879a"
)

func TestHandler_ListPets(t *testing.T) {
	tests := []struct {
		desc         string
		storeError   error
		discordError error
	}{
		{
			desc: "list pets",
		},
		{
			desc:       "store blew up",
			storeError: errors.New("boom"),
		},
		{
			desc:         "discord blew up",
			discordError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := &storeMock{}
			s.On("ListPets").
				Return(store.Pets{{Name: "Chacha"}}, test.storeError).
				Once()

			d := &discordMock{}
			if test.storeError == nil {
				d.On("SendMessage", "Chacha\n").
					Return(&harmony.Message{}, test.discordError).
					Once()
			}

			h := Handler{discord: d, store: s}
			h.ListPets(context.Background())

			s.AssertExpectations(t)
			d.AssertExpectations(t)
		})
	}
}

func TestHandler_Remind(t *testing.T) {
	tests := []struct {
		desc      string
		character string
		pet       store.Pet
	}{
		{
			desc:      "remind chacha on toto",
			character: "Toto",
			pet: store.Pet{
				Name:            "Chacha",
				FoodMinDuration: 1 * time.Hour,
				FoodMaxDuration: 2 * time.Hour,
			},
		},
		{
			desc:      "remind nomoon on titi",
			character: "Titi",
			pet: store.Pet{
				Name:            "Nomoon",
				FoodMinDuration: 5 * time.Hour,
				FoodMaxDuration: 10 * time.Hour,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := &storeMock{}
			s.On("GetPet", test.pet.Name).
				Return(test.pet, nil).
				Once()
			s.On("CreateRemind", mock.MatchedBy(func(r store.Remind) bool {
				return r.PetName == test.pet.Name &&
					r.DiscordUserID == testDiscordUserID &&
					!r.ReminderSent &&
					r.MissedReminder == 0 &&
					r.Character == test.character &&
					r.NextRemind.Sub(time.Now().Add(test.pet.FoodMinDuration)) < time.Second &&
					r.NextRemind.Sub(time.Now().Add(test.pet.FoodMaxDuration)) < time.Second
			})).Return(nil).
				Once()

			r := &reminderMock{}
			r.On("SetUpdate").Return().Once()

			d := &discordMock{}
			d.On("SendMessage", mock.MatchedBy(func(msg string) bool {
				parts := strings.Split(msg, "\n")
				reminderPart := parts[0]
				datePart := parts[1]
				idPart := parts[2]

				if reminderPart != fmt.Sprintf("<@%s> Rappel activé pour familier %q sur %s", testDiscordUserID, test.pet.Name, test.character) {
					return false
				}

				dateParts := strings.SplitN(datePart, ": ", 2)
				date, err := time.Parse(time.RFC1123, dateParts[1])
				if err != nil {
					return false
				}

				if time.Now().Add(test.pet.FoodMinDuration).Sub(date) > time.Minute {
					return false
				}

				id := strings.SplitN(idPart, ": ", 2)[1]

				return id != ""
			})).Return(&harmony.Message{}, nil).Once()

			h := Handler{store: s, discord: d, reminder: r}
			h = setupHandler(t, h)
			h.Remind(context.Background(), &harmony.Message{Content: fmt.Sprintf("!remind %s %s", test.pet.Name, test.character), Author: &harmony.User{ID: testDiscordUserID}})
		})
	}
}

func TestHandler_Remind_validation(t *testing.T) {
	tests := []struct {
		desc    string
		command string
	}{
		{
			desc:    "remind without pet or character",
			command: "!remind",
		},
		{
			desc:    "remind without character",
			command: "!remind Chacha",
		},
		{
			desc:    "remind without character but space",
			command: "!remind Chacha ",
		},
		{
			desc:    "remind without pet but space",
			command: "!remind  Chacha",
		},
		{
			desc:    "remind with additional field",
			command: "!remind Chacha Character toto",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On(
				"SendMessage",
				"Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ").
				Return(&harmony.Message{}, nil).
				Once()

			h := Handler{discord: d}

			message := &harmony.Message{Content: test.command}
			h.Remind(context.Background(), message)

			d.AssertExpectations(t)
		})
	}
}

func TestHandler_Remind_getPetError(t *testing.T) {
	tests := []struct {
		desc             string
		getPetError      error
		sendMessageError error
	}{
		{
			desc:        "get pet not found",
			getPetError: store.NotFoundError{Err: errors.New("not found")},
		},
		{
			desc:        "get pet blew up",
			getPetError: errors.New("boom"),
		},
		{
			desc:             "get pet not found and unable to send message",
			getPetError:      store.NotFoundError{Err: errors.New("not found")},
			sendMessageError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			s := &storeMock{}
			s.On("GetPet", "Chacha").
				Return(store.Pet{}, test.getPetError).
				Once()

			d := &discordMock{}
			if errors.As(test.getPetError, &store.NotFoundError{}) {
				d.On("SendMessage", "\"Chacha\" n'existe pas. `!familiers` pour connaître la liste des familiers gérés.").
					Return(&harmony.Message{}, test.sendMessageError).
					Once()
			}

			h := Handler{store: s, discord: d}
			h = setupHandler(t, h)
			h.Remind(context.Background(), &harmony.Message{Content: "!remind Chacha Test", Author: &harmony.User{ID: testDiscordUserID}})
		})
	}
}

func TestHandler_Remind_createRemindError(t *testing.T) {
	pet := store.Pet{
		Name:            "Chacha",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	s := &storeMock{}
	s.On("GetPet", "Chacha").
		Return(pet, nil).
		Once()
	s.On("CreateRemind", mock.MatchedBy(func(r store.Remind) bool {
		return r.PetName == pet.Name &&
			r.DiscordUserID == testDiscordUserID &&
			!r.ReminderSent &&
			r.MissedReminder == 0 &&
			r.Character == "Test" &&
			r.NextRemind.Sub(time.Now().Add(pet.FoodMinDuration)) < time.Second &&
			r.NextRemind.Sub(time.Now().Add(pet.FoodMaxDuration)) < time.Second
	})).Return(errors.New("boom")).
		Once()

	h := Handler{store: s}
	h = setupHandler(t, h)
	h.Remind(context.Background(), &harmony.Message{Content: "!remind Chacha Test", Author: &harmony.User{ID: testDiscordUserID}})
}

func TestHandler_Remind_sendMessageError(t *testing.T) {
	pet := store.Pet{
		Name:            "Chacha",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	s := &storeMock{}
	s.On("GetPet", "Chacha").
		Return(pet, nil).
		Once()
	s.On("CreateRemind", mock.MatchedBy(func(r store.Remind) bool {
		return r.PetName == pet.Name &&
			r.DiscordUserID == testDiscordUserID &&
			!r.ReminderSent &&
			r.MissedReminder == 0 &&
			r.Character == "Test" &&
			r.NextRemind.Sub(time.Now().Add(pet.FoodMinDuration)) < time.Second &&
			r.NextRemind.Sub(time.Now().Add(pet.FoodMaxDuration)) < time.Second
	})).Return(nil).
		Once()

	r := &reminderMock{}
	r.On("SetUpdate").Return().Once()

	d := &discordMock{}
	d.On("SendMessage", mock.MatchedBy(func(msg string) bool {
		parts := strings.Split(msg, "\n")
		reminderPart := parts[0]
		datePart := parts[1]
		idPart := parts[2]

		if reminderPart != fmt.Sprintf("<@%s> Rappel activé pour familier %q sur %s", testDiscordUserID, "Chacha", "Test") {
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

		id := strings.SplitN(idPart, ": ", 2)[1]

		return id != ""
	})).Return(&harmony.Message{}, errors.New("boom")).Once()

	h := Handler{store: s, reminder: r, discord: d}
	h = setupHandler(t, h)
	h.Remind(context.Background(), &harmony.Message{Content: "!remind Chacha Test", Author: &harmony.User{ID: testDiscordUserID}})
}

func TestHandler_RemoveRemind(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{
		ID:             objectID,
		DiscordUserID:  "2",
		PetName:        "Chacha",
		Character:      "Test",
		MissedReminder: 0,
		NextRemind:     time.Now().Add(1 * time.Hour),
		ReminderSent:   false,
		TimeoutRemind:  time.Now().Add(2 * time.Hour),
	}, nil).Once()

	s.On("RemoveRemind", testRemindID).Return(nil).Once()

	r := &reminderMock{}
	r.On("SetUpdate").Once()

	d := &discordMock{}
	d.On("SendMessage", fmt.Sprintf("<@2> Reminder %q supprimé", testRemindID)).Return(&harmony.Message{}, nil).Once()

	h := Handler{store: s, discord: d, reminder: r}
	h.RemoveRemind(context.Background(), &harmony.Message{Content: "!remove " + testRemindID, Author: &harmony.User{ID: "2"}})
}

func TestHandler_RemoveRemind_validation(t *testing.T) {
	tests := []struct {
		desc             string
		command          string
		wantMessage      string
		sendMessageError error
	}{
		{
			desc:        "bad usage",
			command:     "!remove",
			wantMessage: "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ",
		},
		{
			desc:        "too many arguments",
			command:     "!remove pouet test",
			wantMessage: "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ",
		},
		{
			desc:        "invalid id",
			command:     "!remove 12",
			wantMessage: "<@2> ID invalide.",
		},
		{
			desc:             "invalid id",
			command:          "!remove 12",
			wantMessage:      "<@2> ID invalide.",
			sendMessageError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On("SendMessage", test.wantMessage).Return(&harmony.Message{}, test.sendMessageError).Once()

			h := Handler{discord: d}
			h.RemoveRemind(context.Background(), &harmony.Message{Content: test.command, Author: &harmony.User{ID: "2"}})
		})
	}
}

func TestHandler_RemoveRemind_getRemindError(t *testing.T) {
	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{}, errors.New("boom")).Once()

	h := Handler{store: s}
	h.RemoveRemind(context.Background(), &harmony.Message{Content: "!remove " + testRemindID, Author: &harmony.User{ID: "2"}})
}

func TestHandler_RemoveRemind_badUser(t *testing.T) {
	tests := []struct {
		desc             string
		sendMessageError error
	}{
		{
			desc: "bad user",
		},
		{
			desc:             "bad user",
			sendMessageError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			objectID, err := primitive.ObjectIDFromHex(testRemindID)
			require.NoError(t, err)

			s := &storeMock{}
			s.On("GetRemind", testRemindID).Return(store.Remind{
				ID:             objectID,
				DiscordUserID:  "2",
				PetName:        "Chacha",
				Character:      "Test",
				MissedReminder: 0,
				NextRemind:     time.Now().Add(1 * time.Hour),
				ReminderSent:   false,
				TimeoutRemind:  time.Now().Add(2 * time.Hour),
			}, nil).Once()

			d := &discordMock{}
			d.On("SendMessage", "<@5> Vous ne pouvez pas supprimer un rappel qui ne vous appartient pas.").Return(&harmony.Message{}, test.sendMessageError).Once()

			h := Handler{store: s, discord: d}
			h.RemoveRemind(context.Background(), &harmony.Message{Content: "!remove " + testRemindID, Author: &harmony.User{ID: "5"}})
		})
	}
}

func TestHandler_RemoveRemind_removeRemindError(t *testing.T) {
	tests := []struct {
		desc              string
		removeRemindError error
		sendMessageError  error
	}{
		{
			desc:              "remind not found",
			removeRemindError: store.NotFoundError{Err: errors.New("not found")},
		},
		{
			desc:              "remind not found and send message blew up",
			removeRemindError: store.NotFoundError{Err: errors.New("not found")},
			sendMessageError:  errors.New("boom"),
		},
		{
			desc:              "remove remind blew up",
			removeRemindError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			objectID, err := primitive.ObjectIDFromHex(testRemindID)
			require.NoError(t, err)

			s := &storeMock{}
			s.On("GetRemind", testRemindID).Return(store.Remind{
				ID:             objectID,
				DiscordUserID:  "2",
				PetName:        "Chacha",
				Character:      "Test",
				MissedReminder: 0,
				NextRemind:     time.Now().Add(1 * time.Hour),
				ReminderSent:   false,
				TimeoutRemind:  time.Now().Add(2 * time.Hour),
			}, nil).Once()

			s.On("RemoveRemind", testRemindID).Return(test.removeRemindError).Once()

			d := &discordMock{}
			if errors.As(test.removeRemindError, &store.NotFoundError{}) {
				d.On("SendMessage", fmt.Sprintf("<@2> Pas de rappel avec l'id: %q", testRemindID)).Return(&harmony.Message{}, test.sendMessageError).Once()
			}

			h := Handler{store: s, discord: d}
			h.RemoveRemind(context.Background(), &harmony.Message{Content: "!remove " + testRemindID, Author: &harmony.User{ID: "2"}})
		})
	}
}

func TestHandler_RemoveRemind_sendMessageError(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{
		ID:             objectID,
		DiscordUserID:  "2",
		PetName:        "Chacha",
		Character:      "Test",
		MissedReminder: 0,
		NextRemind:     time.Now().Add(1 * time.Hour),
		ReminderSent:   false,
		TimeoutRemind:  time.Now().Add(2 * time.Hour),
	}, nil).Once()

	s.On("RemoveRemind", testRemindID).Return(nil).Once()

	r := &reminderMock{}
	r.On("SetUpdate").Once()

	d := &discordMock{}
	d.On("SendMessage", fmt.Sprintf("<@2> Reminder %q supprimé", testRemindID)).Return(&harmony.Message{}, errors.New("boom")).Once()

	h := Handler{store: s, discord: d, reminder: r}
	h.RemoveRemind(context.Background(), &harmony.Message{Content: "!remove " + testRemindID, Author: &harmony.User{ID: "2"}})
}

func TestHandler_Help(t *testing.T) {
	tests := []struct {
		desc             string
		sendMessageError error
	}{
		{
			desc: "send help message",
		},
		{
			desc:             "send message error",
			sendMessageError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On("SendMessage", "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ").
				Return(&harmony.Message{}, test.sendMessageError).
				Once()

			h := Handler{discord: d}
			h.Help(context.Background())
		})
	}
}
