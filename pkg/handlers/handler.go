package handlers

import (
	"context"

	"github.com/youkoulayley/reminderbot/pkg/store"
)

// Handler represents a Discord Handler.
type Handler struct {
	store     Storer
	reminder  Reminder
	channelID string
}

// NewHandler creates a new Handler.
func NewHandler(s Storer, r Reminder, channelID string) Handler {
	return Handler{
		store:     s,
		reminder:  r,
		channelID: channelID,
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
