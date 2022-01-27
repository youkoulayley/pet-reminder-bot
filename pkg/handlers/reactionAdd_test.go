package handlers

import (
	"testing"

	"github.com/skwair/harmony"
	"github.com/skwair/harmony/discord"
	"github.com/youkoulayley/pet-reminder-bot/pkg/bot"
)

func TestHandler_ReactionAdd(t *testing.T) {
	b := &botMock{}
	b.On("NewCycle", bot.NewCycleConfig{
		AuthorID:  "3",
		MessageID: "123",
	}).Once()

	h := Handler{
		bot:     b,
		botUser: discord.User{ID: "3"},
	}

	msg := &harmony.MessageReaction{MessageID: "123", UserID: "3"}
	h.ReactionAdd(msg)

	b.AssertExpectations(t)
}
