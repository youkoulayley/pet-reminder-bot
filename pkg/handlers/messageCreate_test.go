package handlers

import (
	"testing"

	"github.com/skwair/harmony/discord"
	"github.com/youkoulayley/pet-reminder-bot/pkg/bot"
)

func TestHandler_MessageCreate_botMessage(t *testing.T) {
	b := &botMock{}
	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "3"},
	}

	msg := &discord.Message{Content: "!help", Author: discord.User{ID: "3"}}
	h.MessageCreate(msg)

	b.AssertExpectations(t)
}

func TestHandler_MessageCreate_unknownCommand(t *testing.T) {
	b := &botMock{}
	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "2"},
	}

	msg := &discord.Message{Content: "!pouet", Author: discord.User{ID: "3"}}
	h.MessageCreate(msg)

	b.AssertExpectations(t)
}

func TestHandler_MessageCreate_listPetsCommand(t *testing.T) {
	b := &botMock{}
	b.On("ListPets").Once()

	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "2"},
	}

	msg := &discord.Message{Content: "!familiers", Author: discord.User{ID: "3"}}
	h.MessageCreate(msg)

	b.AssertExpectations(t)
}

func TestHandler_MessageCreate_helpCommand(t *testing.T) {
	b := &botMock{}
	b.On("Help").Once()

	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "2"},
	}

	msg := &discord.Message{Content: "!help", Author: discord.User{ID: "3"}}
	h.MessageCreate(msg)

	b.AssertExpectations(t)
}

func TestHandler_MessageCreate_remindCommand_validation(t *testing.T) {
	tests := []struct {
		desc    string
		command string
	}{
		{
			desc:    "command invalid",
			command: "!remind",
		},
		{
			desc:    "pet empty",
			command: "!remind  toto",
		},
		{
			desc:    "character empty",
			command: "!remind Chacha ",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			b := &botMock{}
			b.On("Help").Once()

			h := Handler{
				bot:     b,
				botUser: discord.User{ID: "2"},
			}

			msg := &discord.Message{Content: test.command, Author: discord.User{ID: "3"}}
			h.MessageCreate(msg)

			b.AssertExpectations(t)
		})
	}
}

func TestHandler_MessageCreate_remindCommand(t *testing.T) {
	b := &botMock{}
	b.On("Remind", bot.RemindConfig{
		AuthorID:  "3",
		Pet:       "Chacha",
		Character: "Toto",
	}).Once()

	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "2"},
	}

	msg := &discord.Message{Content: "!remind Chacha Toto", Author: discord.User{ID: "3"}}
	h.MessageCreate(msg)

	b.AssertExpectations(t)
}

func TestHandler_MessageCreate_removeCommand_validation(t *testing.T) {
	tests := []struct {
		desc    string
		command string
	}{
		{
			desc:    "command invalid",
			command: "!remove",
		},
		{
			desc:    "id empty",
			command: "!remove ",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			b := &botMock{}
			b.On("Help").Once()

			h := Handler{
				bot:     b,
				botUser: discord.User{ID: "2"},
			}

			msg := &discord.Message{Content: test.command, Author: discord.User{ID: "3"}}
			h.MessageCreate(msg)

			b.AssertExpectations(t)
		})
	}
}

func TestHandler_MessageCreate_removeCommand(t *testing.T) {
	b := &botMock{}
	b.On("RemoveRemind", bot.RemoveRemindConfig{
		AuthorID: "3",
		ID:       "123",
	}).Once()

	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "2"},
	}

	msg := &discord.Message{Content: "!remove 123", Author: discord.User{ID: "3"}}
	h.MessageCreate(msg)

	b.AssertExpectations(t)
}
