package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const helpMessage = `Commandes disponible:
  - ` + "`!familiers`" + `
  - ` + "`!list`" + `
  - ` + "`!remind <Familier> <Personnage>`" + `
  - ` + "`!remove <ID>` "

// ListPets handles the familiers command for the bot.
// Call it with `!familiers`.
func (b *Bot) ListPets(ctx context.Context) {
	pets, err := b.store.ListPets(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Unable to list pets")

		return
	}

	if _, err = b.discord.SendMessage(ctx, pets.String()); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// RemindConfig represents remind command config.
type RemindConfig struct {
	AuthorID  string
	Pet       string
	Character string
}

// Validate ensures that all fields are valid.
func (c RemindConfig) Validate() error {
	if c.AuthorID == "" {
		return errors.New("author id cannot be empty")
	}

	if c.Pet == "" {
		return errors.New("pet cannot be empty")
	}

	if c.Character == "" {
		return errors.New("character cannot be empty")
	}

	return nil
}

// Remind handles the remind command for the bot.
// Call it with `!remind <PetName> <CharacterName>`
// PetName can be found with the ListPets command.
func (b *Bot) Remind(ctx context.Context, cfg RemindConfig) {
	if err := cfg.Validate(); err != nil {
		b.Help(ctx)

		return
	}

	petDuration, err := b.store.GetPet(ctx, cfg.Pet)
	if err != nil {
		if errors.As(err, &store.NotFoundError{}) {
			message := fmt.Sprintf("%q n'existe pas. `!familiers` pour connaître la liste des familiers gérés.", cfg.Pet)
			if _, err = b.discord.SendMessage(ctx, message); err != nil {
				log.Error().Err(err).Msg("Unable to send message")

				return
			}

			return
		}

		log.Error().Err(err).Msg("Unable to get pet")

		return
	}

	id := primitive.NewObjectID()
	remind := store.Remind{
		ID:            id,
		DiscordUserID: cfg.AuthorID,
		PetName:       cfg.Pet,
		Character:     cfg.Character,
		NextRemind:    time.Now().Add(petDuration.FoodMinDuration),
		TimeoutRemind: time.Now().Add(petDuration.FoodMaxDuration),
	}

	if err = b.store.CreateRemind(ctx, remind); err != nil {
		log.Error().Err(err).Msg("Unable to create reminder")

		return
	}

	b.reminder.SetUpdate()

	message := fmt.Sprintf(
		"<@%s> Rappel activé pour familier %q sur %s\nProchain rappel: %s\nID: %s\n",
		cfg.AuthorID,
		cfg.Pet,
		cfg.Character,
		remind.NextRemind.In(b.timezone).Format(time.RFC1123),
		id.Hex(),
	)
	if _, err = b.discord.SendMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// RemoveRemindConfig represents remove remind command config.
type RemoveRemindConfig struct {
	AuthorID string
	ID       string
}

// Validate ensures that all fields are valid.
func (c RemoveRemindConfig) Validate() error {
	if c.AuthorID == "" {
		return errors.New("author id cannot be empty")
	}

	if _, err := primitive.ObjectIDFromHex(c.ID); err != nil {
		return fmt.Errorf("object id from hex: %w", err)
	}

	return nil
}

// RemoveRemind handles the remove command for the bot.
// Call it with `!remove <RemindID>`.
func (b *Bot) RemoveRemind(ctx context.Context, cfg RemoveRemindConfig) {
	if err := cfg.Validate(); err != nil {
		b.Help(ctx)

		return
	}

	logger := log.With().Str("id", cfg.ID).Logger()

	remind, err := b.store.GetRemind(ctx, cfg.ID)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to find remind")

		return
	}

	if remind.DiscordUserID != cfg.AuthorID {
		message := fmt.Sprintf("<@%s> Vous ne pouvez pas supprimer un rappel qui ne vous appartient pas.", cfg.AuthorID)
		if _, err = b.discord.SendMessage(ctx, message); err != nil {
			logger.Error().Err(err).Msg("Unable to send message")

			return
		}

		logger.Debug().Msg("Unable to disable reminder: wrong discordUserID")

		return
	}

	if err = b.store.RemoveRemind(ctx, cfg.ID); err != nil {
		if errors.As(err, &store.NotFoundError{}) {
			logger.Debug().Err(err).Msg("Unable to find reminder")

			message := fmt.Sprintf("<@%s> Pas de rappel avec l'id: %q", cfg.AuthorID, cfg.ID)
			if _, err = b.discord.SendMessage(ctx, message); err != nil {
				logger.Error().Err(err).Msg("Unable to send message")

				return
			}

			return
		}

		logger.Error().Err(err).Msg("Unable to remove reminder")

		return
	}

	b.reminder.SetUpdate()

	message := fmt.Sprintf("<@%s> Rappel %q supprimé", cfg.AuthorID, cfg.ID)
	if _, err = b.discord.SendMessage(ctx, message); err != nil {
		logger.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// Help handles all other commands.
func (b *Bot) Help(ctx context.Context) {
	if _, err := b.discord.SendMessage(ctx, helpMessage); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// NewCycleConfig represents new cycle config.
type NewCycleConfig struct {
	AuthorID  string
	MessageID string
}

// Validate ensures that all fields are valid.
func (c *NewCycleConfig) Validate() error {
	if c.AuthorID == "" {
		return errors.New("author id cannot be empty")
	}

	if c.MessageID == "" {
		return errors.New("message id cannot be empty")
	}

	return nil
}

// NewCycle starts a new cycle when the user add a reaction to a message.
func (b *Bot) NewCycle(ctx context.Context, cfg NewCycleConfig) {
	if err := cfg.Validate(); err != nil {
		b.Help(ctx)

		return
	}

	message, err := b.discord.Message(ctx, cfg.MessageID)
	if err != nil {
		log.Error().Err(err).Msg("Unable to find message")

		return
	}

	parts := strings.Split(message.Content, "ID:")
	if len(parts) != 2 {
		log.Debug().Msg("ID invalid")

		return
	}

	id := strings.TrimSpace(parts[1])
	if _, err = primitive.ObjectIDFromHex(id); err != nil {
		log.Debug().Msg("ID invalid")

		return
	}

	remind, err := b.store.GetRemind(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Unable to get remind")

		return
	}

	if remind.DiscordUserID != cfg.AuthorID {
		log.Error().Msg("Invalid user")

		return
	}

	pet, err := b.store.GetPet(ctx, remind.PetName)
	if err != nil {
		log.Error().Err(err).Msg("Invalid pet")

		return
	}

	remind.MissedReminder = 0
	remind.ReminderSent = false
	remind.NextRemind = time.Now().Add(pet.FoodMinDuration)
	remind.TimeoutRemind = time.Now().Add(pet.FoodMaxDuration)

	if err = b.store.UpdateRemind(ctx, remind); err != nil {
		log.Error().Err(err).Msg("Unable to update remind")

		return
	}

	b.reminder.SetUpdate()
}

// ListReminds lists all reminds set for the user identified by the given id.
func (b *Bot) ListReminds(ctx context.Context, id string) {
	logger := log.With().Str("id", id).Logger()

	reminds, err := b.store.ListRemindsByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to list reminds")

		return
	}

	if len(reminds) == 0 {
		if _, err = b.discord.SendMessage(ctx, fmt.Sprintf("<@%s> Aucun rappel disponible", id)); err != nil {
			logger.Error().Err(err).Msg("Unable to send message")

			return
		}

		return
	}

	message := []string{fmt.Sprintf("<@%s> Liste de vos rappels:", id)}

	for _, remind := range reminds {
		r := fmt.Sprintf("  - %s - %s sur %s - Prochain rappel: %s", remind.ID.Hex(), remind.PetName, remind.Character, remind.NextRemind.In(b.timezone).Format(time.RFC1123))
		message = append(message, r)
	}

	if _, err = b.discord.SendMessage(ctx, strings.Join(message, "\n")); err != nil {
		logger.Error().Err(err).Msg("Unable to send message")

		return
	}
}
