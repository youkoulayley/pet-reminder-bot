package reminder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.uber.org/atomic"
)

// Storer is capable of interacting with the store.
type Storer interface {
	GetPet(ctx context.Context, name string) (store.Pet, error)

	ListReminds(ctx context.Context) ([]store.Remind, error)
	UpdateRemind(ctx context.Context, remind store.Remind) error
}

// Discord is capable of interacting with discord.
type Discord interface {
	SendMessage(ctx context.Context, text string) (*harmony.Message, error)
}

// Reminder represents the reminder.
type Reminder struct {
	reminds   []store.Remind
	remindsMu sync.Mutex

	needUpdate *atomic.Bool

	store   Storer
	discord Discord
}

// New creates a new Reminder.
func New(s Storer, d Discord) (*Reminder, error) {
	reminder := Reminder{
		store:      s,
		discord:    d,
		needUpdate: atomic.NewBool(true),
	}

	return &reminder, nil
}

// SetUpdate notify the reminder to load the new reminds.
func (r *Reminder) SetUpdate() {
	r.needUpdate.Store(true)
}

// LoadReminds loads the reminds stored in the storer and load it to the memory.
func (r *Reminder) LoadReminds(ctx context.Context) error {
	reminds, err := r.store.ListReminds(ctx)
	if err != nil {
		return fmt.Errorf("list reminds: %w", err)
	}

	r.remindsMu.Lock()
	r.reminds = reminds
	r.remindsMu.Unlock()

	return nil
}

// Run starts the ticker.
func (r *Reminder) Run(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-t.C:
			r.Process(ctx)
		}
	}
}

// Process handles the reminder logic.
func (r *Reminder) Process(ctx context.Context) {
	if r.needUpdate.Load() {
		if err := r.LoadReminds(ctx); err != nil {
			log.Error().Err(err).Msg("Unable to load reminder")

			return
		}

		r.needUpdate.Store(false)
	}

	var needUpdate bool

	for _, remind := range r.reminds {
		if remind.NextRemind.Before(time.Now()) && !remind.ReminderSent {
			message := fmt.Sprintf("<@%s> Il faut nourrir %q sur %s\nID: %s", remind.DiscordUserID, remind.PetName, remind.Character, remind.ID.Hex())
			if _, err := r.discord.SendMessage(ctx, message); err != nil {
				log.Error().Err(err).Msg("Unable to send reminder message")

				continue
			}

			remind.ReminderSent = true
			if err := r.store.UpdateRemind(ctx, remind); err != nil {
				log.Error().Err(err).Msg("Unable to update remind")

				continue
			}

			needUpdate = true
		}

		if remind.TimeoutRemind.Before(time.Now()) && remind.ReminderSent {
			pet, err := r.store.GetPet(ctx, remind.PetName)
			if err != nil {
				log.Error().Err(err).Msg("Unable to get pet")

				continue
			}

			remind.ReminderSent = false
			remind.MissedReminder++
			remind.NextRemind = time.Now().Add(pet.FoodMinDuration)
			remind.TimeoutRemind = time.Now().Add(pet.FoodMaxDuration)

			if err = r.store.UpdateRemind(ctx, remind); err != nil {
				log.Error().Err(err).Msg("Unable to update remind")

				continue
			}

			needUpdate = true

			message := fmt.Sprintf("<@%s> %q sur %s a râté %d repas.\nProchain rappel: %s\nID: %s", remind.DiscordUserID, remind.PetName, remind.Character, remind.MissedReminder, remind.NextRemind.Format(time.RFC1123), remind.ID.Hex())
			if _, err = r.discord.SendMessage(ctx, message); err != nil {
				log.Error().Err(err).Msg("Unable to send reminder message")

				continue
			}
		}
	}

	if needUpdate {
		if err := r.LoadReminds(ctx); err != nil {
			log.Error().Err(err).Msg("Unable to load remind")
		}
	}
}
