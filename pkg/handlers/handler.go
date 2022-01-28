package handlers

import (
	"context"

	"github.com/skwair/harmony/discord"
	"github.com/youkoulayley/pet-reminder-bot/pkg/bot"
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
	ListReminds(ctx context.Context, id string)
	Remind(ctx context.Context, cfg bot.RemindConfig)
	RemoveRemind(ctx context.Context, cfg bot.RemoveRemindConfig)
	Help(ctx context.Context)
	NewCycle(ctx context.Context, cfg bot.NewCycleConfig)
}
