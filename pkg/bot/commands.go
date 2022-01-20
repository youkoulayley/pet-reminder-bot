package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony"
	"github.com/skwair/harmony/discord"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// Remind handles the remind command for the bot.
// Call it with `!remind <PetName> <CharacterName>`
// PetName can be found with the ListPets command.
func (b *Bot) Remind(ctx context.Context, m *discord.Message) {
	parts := strings.Split(m.Content, " ")
	if len(parts) != 3 {
		b.Help(ctx)

		return
	}

	pet := parts[1]
	character := parts[2]

	if pet == "" || character == "" {
		b.Help(ctx)

		return
	}

	petDuration, err := b.store.GetPet(ctx, pet)
	if err != nil {
		if errors.As(err, &store.NotFoundError{}) {
			message := fmt.Sprintf("%q n'existe pas. `!familiers` pour connaître la liste des familiers gérés.", pet)
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
		DiscordUserID: m.Author.ID,
		PetName:       pet,
		Character:     character,
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
		m.Author.ID,
		pet,
		character,
		remind.NextRemind.In(b.timezone).Format(time.RFC1123),
		id.Hex(),
	)
	if _, err = b.discord.SendMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// RemoveRemind handles the remove command for the bot.
// Call it with `!remove <RemindID>`.
func (b *Bot) RemoveRemind(ctx context.Context, m *discord.Message) {
	parts := strings.Split(m.Content, " ")
	if len(parts) != 2 {
		b.Help(ctx)

		return
	}

	id := parts[1]
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		log.Debug().Msg("ID invalid")

		message := fmt.Sprintf("<@%s> ID invalide.", m.Author.ID)
		if _, err := b.discord.SendMessage(ctx, message); err != nil {
			log.Error().Err(err).Msg("Unable to send message")

			return
		}

		return
	}

	logger := log.With().Str("id", id).Logger()

	remind, err := b.store.GetRemind(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to find remind")

		return
	}

	if remind.DiscordUserID != m.Author.ID {
		message := fmt.Sprintf("<@%s> Vous ne pouvez pas supprimer un rappel qui ne vous appartient pas.", m.Author.ID)
		if _, err = b.discord.SendMessage(ctx, message); err != nil {
			logger.Error().Err(err).Msg("Unable to send message")

			return
		}

		logger.Debug().Msg("Unable to disable reminder: wrong discordUserID")

		return
	}

	if err = b.store.RemoveRemind(ctx, id); err != nil {
		if errors.As(err, &store.NotFoundError{}) {
			logger.Debug().Err(err).Msg("Unable to find reminder")

			message := fmt.Sprintf("<@%s> Pas de rappel avec l'id: %q", m.Author.ID, id)
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

	message := fmt.Sprintf("<@%s> Rappel %q supprimé", m.Author.ID, id)
	if _, err = b.discord.SendMessage(ctx, message); err != nil {
		logger.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// Help handles all other commands.
func (b *Bot) Help(ctx context.Context) {
	message := "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` "
	if _, err := b.discord.SendMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// NewCycle starts a new cycle when the user add a reaction to a message.
func (b *Bot) NewCycle(ctx context.Context, m *harmony.MessageReaction) {
	message, err := b.discord.Message(ctx, m.MessageID)
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

	if remind.DiscordUserID != m.UserID {
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
