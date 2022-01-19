package handlers

import (
	"context"
	"time"

	"github.com/skwair/harmony"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
)

// Handler represents a Discord Handler.
type Handler struct {
	bot      *harmony.User
	discord  Discord
	store    Storer
	reminder Reminder

	timezone *time.Location
}

// NewHandler creates a new Handler.
func NewHandler(b *harmony.User, d Discord, s Storer, r Reminder, tz *time.Location) Handler {
	return Handler{
		bot:      b,
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
	Message(ctx context.Context, id string) (*harmony.Message, error)
	SendMessage(ctx context.Context, text string) (*harmony.Message, error)
}
