package handlers

import (
	"context"

	"github.com/skwair/harmony"
	"github.com/skwair/harmony/discord"
)

// Handler represents a Discord Handler.
type Handler struct {
	bot     Bot
	botUser discord.User
}

// New creates a new Handler.
func New(b Bot, bu discord.User) Handler {
	return Handler{
		bot:     b,
		botUser: bu,
	}
}

// Bot is capable of interacting with the bot.
type Bot interface {
	ListPets(ctx context.Context)
	Remind(ctx context.Context, m *discord.Message)
	RemoveRemind(ctx context.Context, m *discord.Message)
	Help(ctx context.Context)
	NewCycle(ctx context.Context, m *harmony.MessageReaction)
}
