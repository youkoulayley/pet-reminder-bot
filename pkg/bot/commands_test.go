package bot

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/skwair/harmony/discord"
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
					Return(&discord.Message{}, test.discordError).
					Once()
			}

			b := Bot{discord: d, store: s}
			b.ListPets(context.Background())

			s.AssertExpectations(t)
			d.AssertExpectations(t)
		})
	}
}

func TestHandler_Remind(t *testing.T) {
	tests := []struct {
		desc   string
		config RemindConfig
		pet    store.Pet
	}{
		{
			desc: "remind chacha on toto",
			config: RemindConfig{
				AuthorID:  testDiscordUserID,
				Pet:       "Chacha",
				Character: "Toto",
			},
			pet: store.Pet{
				Name:            "Chacha",
				FoodMinDuration: 1 * time.Hour,
				FoodMaxDuration: 2 * time.Hour,
			},
		},
		{
			desc: "remind nomoon on titi",
			config: RemindConfig{
				AuthorID:  testDiscordUserID,
				Pet:       "Nomoon",
				Character: "Titi",
			},
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
			s.On("GetPet", test.config.Pet).
				Return(test.pet, nil).
				Once()
			s.On("CreateRemind", mock.MatchedBy(func(r store.Remind) bool {
				return r.PetName == test.pet.Name &&
					r.DiscordUserID == test.config.AuthorID &&
					!r.ReminderSent &&
					r.MissedReminder == 0 &&
					r.Character == test.config.Character &&
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

				if reminderPart != fmt.Sprintf("<@%s> Rappel activé pour familier %q sur %s", testDiscordUserID, test.pet.Name, test.config.Character) {
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
			})).Return(&discord.Message{}, nil).Once()

			b := Bot{store: s, discord: d, reminder: r}
			b = setupBot(t, b)
			b.Remind(context.Background(), test.config)
		})
	}
}

func TestHandler_Remind_validation(t *testing.T) {
	tests := []struct {
		desc   string
		config RemindConfig
	}{
		{
			desc:   "remind without pet or character or author id",
			config: RemindConfig{},
		},
		{
			desc: "remind without author ID",
			config: RemindConfig{
				Pet:       "Chacha",
				Character: "Toto",
			},
		},
		{
			desc: "remind without character",
			config: RemindConfig{
				AuthorID: testDiscordUserID,
				Pet:      "Chacha",
			},
		},
		{
			desc: "remind without pet",
			config: RemindConfig{
				AuthorID:  testDiscordUserID,
				Character: "Toto",
			},
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
				Return(&discord.Message{}, nil).
				Once()

			b := Bot{discord: d}

			b.Remind(context.Background(), test.config)

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
					Return(&discord.Message{}, test.sendMessageError).
					Once()
			}

			b := Bot{store: s, discord: d}
			b = setupBot(t, b)

			cfg := RemindConfig{
				AuthorID:  testDiscordUserID,
				Pet:       "Chacha",
				Character: "Test",
			}
			b.Remind(context.Background(), cfg)
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

	b := Bot{store: s}
	b = setupBot(t, b)

	cfg := RemindConfig{
		AuthorID:  testDiscordUserID,
		Pet:       "Chacha",
		Character: "Test",
	}
	b.Remind(context.Background(), cfg)
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
	})).Return(&discord.Message{}, errors.New("boom")).Once()

	b := Bot{store: s, reminder: r, discord: d}
	b = setupBot(t, b)

	cfg := RemindConfig{
		AuthorID:  testDiscordUserID,
		Pet:       "Chacha",
		Character: "Test",
	}
	b.Remind(context.Background(), cfg)
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
	d.On("SendMessage", fmt.Sprintf("<@2> Rappel %q supprimé", testRemindID)).Return(&discord.Message{}, nil).Once()

	b := Bot{store: s, discord: d, reminder: r}
	b.RemoveRemind(context.Background(), RemoveRemindConfig{
		AuthorID: testDiscordUserID,
		ID:       testRemindID,
	})
}

func TestHandler_RemoveRemind_validation(t *testing.T) {
	tests := []struct {
		desc             string
		config           RemoveRemindConfig
		wantMessage      string
		sendMessageError error
	}{
		{
			desc:        "bad usage",
			config:      RemoveRemindConfig{},
			wantMessage: "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ",
		},
		{
			desc: "emty author id",
			config: RemoveRemindConfig{
				ID: testRemindID,
			},
			wantMessage: "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ",
		},
		{
			desc: "invalid id",
			config: RemoveRemindConfig{
				AuthorID: testDiscordUserID,
				ID:       "12",
			},
			wantMessage: "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ",
		},
		{
			desc: "send message blew up",
			config: RemoveRemindConfig{
				AuthorID: testDiscordUserID,
				ID:       "12",
			},
			wantMessage:      "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ",
			sendMessageError: errors.New("boom"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On("SendMessage", test.wantMessage).Return(&discord.Message{}, test.sendMessageError).Once()

			b := Bot{discord: d}
			b.RemoveRemind(context.Background(), test.config)
		})
	}
}

func TestHandler_RemoveRemind_getRemindError(t *testing.T) {
	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{}, errors.New("boom")).Once()

	b := Bot{store: s}
	cfg := RemoveRemindConfig{
		AuthorID: testDiscordUserID,
		ID:       testRemindID,
	}
	b.RemoveRemind(context.Background(), cfg)
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
			d.On("SendMessage", "<@5> Vous ne pouvez pas supprimer un rappel qui ne vous appartient pas.").Return(&discord.Message{}, test.sendMessageError).Once()

			b := Bot{store: s, discord: d}

			cfg := RemoveRemindConfig{
				AuthorID: "5",
				ID:       testRemindID,
			}
			b.RemoveRemind(context.Background(), cfg)
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
				d.On("SendMessage", fmt.Sprintf("<@2> Pas de rappel avec l'id: %q", testRemindID)).Return(&discord.Message{}, test.sendMessageError).Once()
			}

			b := Bot{store: s, discord: d}

			cfg := RemoveRemindConfig{
				AuthorID: testDiscordUserID,
				ID:       testRemindID,
			}
			b.RemoveRemind(context.Background(), cfg)
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
	d.On("SendMessage", fmt.Sprintf("<@2> Rappel %q supprimé", testRemindID)).Return(&discord.Message{}, errors.New("boom")).Once()

	b := Bot{store: s, discord: d, reminder: r}

	cfg := RemoveRemindConfig{
		AuthorID: testDiscordUserID,
		ID:       testRemindID,
	}
	b.RemoveRemind(context.Background(), cfg)
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
				Return(&discord.Message{}, test.sendMessageError).
				Once()

			b := Bot{discord: d}
			b.Help(context.Background())
		})
	}
}

func TestHandler_NewCycle(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	pet := store.Pet{
		Name:            "Chacha",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	d := &discordMock{}
	d.On("Message", "123").Return(&discord.Message{Content: "ID: " + testRemindID}, nil).Once()

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

	b := Bot{store: s, reminder: r, discord: d}

	cfg := NewCycleConfig{
		AuthorID:  testDiscordUserID,
		MessageID: "123",
	}
	b.NewCycle(context.Background(), cfg)
}

func TestHandler_NewCycle_messageError(t *testing.T) {
	d := &discordMock{}
	d.On("Message", "123").Return(&discord.Message{}, errors.New("boom")).Once()

	b := Bot{discord: d}

	cfg := NewCycleConfig{
		AuthorID:  testDiscordUserID,
		MessageID: "123",
	}
	b.NewCycle(context.Background(), cfg)
}

func TestHandler_NewCycle_validation(t *testing.T) {
	tests := []struct {
		desc   string
		config NewCycleConfig
	}{
		{
			desc:   "message id empty",
			config: NewCycleConfig{AuthorID: "5"},
		},
		{
			desc:   "author id empty",
			config: NewCycleConfig{MessageID: "123"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On("SendMessage", "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` ").
				Return(&discord.Message{}, nil).
				Once()

			b := Bot{discord: d}
			b.NewCycle(context.Background(), test.config)

			d.AssertExpectations(t)
		})
	}
}

func TestHandler_NewCycle_validation_messageID(t *testing.T) {
	tests := []struct {
		desc    string
		message string
	}{
		{
			desc:    "no id",
			message: "no id",
		},
		{
			desc:    "bad id",
			message: "ID:",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			d := &discordMock{}
			d.On("Message", "123").Return(&discord.Message{Content: test.message}, nil).Once()

			b := Bot{discord: d}
			cfg := NewCycleConfig{
				AuthorID:  "5",
				MessageID: "123",
			}
			b.NewCycle(context.Background(), cfg)

			d.AssertExpectations(t)
		})
	}
}

func TestHandler_NewCycle_getRemindError(t *testing.T) {
	d := &discordMock{}
	d.On("Message", "123").Return(&discord.Message{Content: "ID: " + testRemindID}, nil).Once()

	s := &storeMock{}
	s.On("GetRemind", testRemindID).Return(store.Remind{}, errors.New("boom")).Once()

	b := Bot{store: s, discord: d}

	cfg := NewCycleConfig{
		AuthorID:  testDiscordUserID,
		MessageID: "123",
	}
	b.NewCycle(context.Background(), cfg)
}

func TestHandler_NewCycle_badUser(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	d := &discordMock{}
	d.On("Message", "123").Return(&discord.Message{Content: "ID: " + testRemindID}, nil).Once()

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

	b := Bot{store: s, discord: d}

	cfg := NewCycleConfig{
		AuthorID:  testDiscordUserID,
		MessageID: "123",
	}
	b.NewCycle(context.Background(), cfg)
}

func TestHandler_NewCycle_getPetError(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	d := &discordMock{}
	d.On("Message", "123").Return(&discord.Message{Content: "ID: " + testRemindID}, nil).Once()

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

	b := Bot{store: s, discord: d}

	cfg := NewCycleConfig{
		AuthorID:  testDiscordUserID,
		MessageID: "123",
	}
	b.NewCycle(context.Background(), cfg)
}

func TestHandler_NewCycle_updateRemindError(t *testing.T) {
	objectID, err := primitive.ObjectIDFromHex(testRemindID)
	require.NoError(t, err)

	pet := store.Pet{
		Name:            "Chacha",
		FoodMinDuration: 1 * time.Hour,
		FoodMaxDuration: 2 * time.Hour,
	}

	d := &discordMock{}
	d.On("Message", "123").Return(&discord.Message{Content: "ID: " + testRemindID}, nil).Once()

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

	b := Bot{store: s, discord: d}

	cfg := NewCycleConfig{
		AuthorID:  testDiscordUserID,
		MessageID: "123",
	}
	b.NewCycle(context.Background(), cfg)
}
