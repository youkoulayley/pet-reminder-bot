package reminder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/youkoulayley/reminderbot/pkg/store"
)

// Storer is capable of interacting with the store.
type Storer interface {
	GetPet(ctx context.Context, name string) (store.Pet, error)

	ListReminds(ctx context.Context) ([]store.Remind, error)
	UpdateRemind(ctx context.Context, remind store.Remind) error
}

// Reminder represents the reminder.
type Reminder struct {
	reminds   []store.Remind
	remindsMu sync.Mutex

	needUpdate bool

	discord          *discordgo.Session
	discordChannelID string

	store Storer
}

// New creates a new Reminder.
func New(s Storer, d *discordgo.Session, discordChannelID string) (*Reminder, error) {
	reminder := Reminder{
		store:            s,
		discord:          d,
		discordChannelID: discordChannelID,
	}

	err := reminder.LoadReminds(context.Background())
	if err != nil {
		return nil, err
	}

	return &reminder, nil
}

// SetUpdate notify the reminder to load the new reminds.
func (r *Reminder) SetUpdate() {
	r.needUpdate = true
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
	if r.needUpdate {
		if err := r.LoadReminds(ctx); err != nil {
			log.Error().Err(err).Msg("Load reminder")
			return
		}

		r.needUpdate = false
	}

	var needUpdate bool

	for _, remind := range r.reminds {
		if remind.NextRemind.Before(time.Now()) && !remind.ReminderSent {
			if _, err := r.discord.ChannelMessageSend(r.discordChannelID, fmt.Sprintf("<@%s> Il faut nourrir %q sur %s\nID: %s", remind.DiscordUserID, remind.PetName, remind.Character, remind.ID.Hex())); err != nil {
				log.Error().Err(err).Msg("Unable to send reminder message")
				continue
			}

			remind.ReminderSent = true
			if err := r.store.UpdateRemind(ctx, remind); err != nil {
				log.Error().Err(err).Msg("Update remind")
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
				log.Error().Err(err).Msg("Update remind")
				continue
			}

			if _, err = r.discord.ChannelMessageSend(r.discordChannelID, fmt.Sprintf("<@%s> %q sur %s a râté %d repas.\nProchain rappel: %s\nID: %s", remind.DiscordUserID, remind.PetName, remind.Character, remind.MissedReminder, remind.NextRemind.Format(time.RFC1123), remind.ID.Hex())); err != nil {
				log.Error().Err(err).Msg("Unable to send reminder message")
				continue
			}

			needUpdate = true
		}
	}

	if needUpdate {
		if err := r.LoadReminds(ctx); err != nil {
			log.Error().Err(err).Msg("Unable to load remind")
		}
	}
}
