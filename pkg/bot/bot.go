package bot

import (
	"context"
	"time"

	"github.com/skwair/harmony/discord"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
)

// Bot represents the Discord bot.
type Bot struct {
	discord  Discord
	store    Storer
	reminder Reminder

	timezone *time.Location
}

// New creates a bot.
func New(d Discord, s Storer, r Reminder, tz *time.Location) *Bot {
	return &Bot{
		discord:  d,
		store:    s,
		reminder: r,
		timezone: tz,
	}
}

// Storer is capable of interacting with the store.
type Storer interface {
	ListPets(ctx context.Context) (store.Pets, error)
	GetPet(ctx context.Context, name string) (store.Pet, error)
	CreateRemind(ctx context.Context, remind store.Remind) error
	UpdateRemind(ctx context.Context, remind store.Remind) error
	GetRemind(ctx context.Context, id string) (store.Remind, error)
	RemoveRemind(ctx context.Context, id string) error
}

// Reminder is capable of interacting with the reminder.
type Reminder interface {
	SetUpdate()
}

// Discord is capable of interacting with Discord.
type Discord interface {
	Message(ctx context.Context, id string) (*discord.Message, error)
	SendMessage(ctx context.Context, text string) (*discord.Message, error)
}
