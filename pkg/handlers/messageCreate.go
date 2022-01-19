package handlers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageCreate gets all message created.
// All messages send by the bot are ignored.
func (h *Handler) MessageCreate(m *harmony.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if h.bot.ID == m.Author.ID {
		log.Debug().Msg("Skipping message sent by me")

		return
	}

	switch {
	case strings.HasPrefix(m.Content, "!familiers"):
		h.ListPets(ctx)
	case strings.HasPrefix(m.Content, "!remind"):
		h.Remind(ctx, m)
	case strings.HasPrefix(m.Content, "!remove"):
		h.RemoveRemind(ctx, m)
	case strings.HasPrefix(m.Content, "!help"):
		h.Help(ctx)
	default:
	}
}

// ListPets handles the familiers command for the bot.
// Call it with `!familiers`.
func (h *Handler) ListPets(ctx context.Context) {
	pets, err := h.store.ListPets(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Unable to list pets")

		return
	}

	if _, err = h.discord.SendMessage(ctx, pets.String()); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// Remind handles the remind command for the bot.
// Call it with `!remind <PetName> <CharacterName>`
// PetName can be found with the ListPets command.
func (h *Handler) Remind(ctx context.Context, m *harmony.Message) {
	parts := strings.Split(m.Content, " ")
	if len(parts) != 3 {
		h.Help(ctx)

		return
	}

	pet := parts[1]
	character := parts[2]

	if pet == "" || character == "" {
		h.Help(ctx)

		return
	}

	petDuration, err := h.store.GetPet(ctx, pet)
	if err != nil {
		if errors.As(err, &store.NotFoundError{}) {
			message := fmt.Sprintf("%q n'existe pas. `!familiers` pour connaître la liste des familiers gérés.", pet)
			if _, err = h.discord.SendMessage(ctx, message); err != nil {
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

	if err = h.store.CreateRemind(ctx, remind); err != nil {
		log.Error().Err(err).Msg("Unable to create reminder")

		return
	}

	h.reminder.SetUpdate()

	message := fmt.Sprintf(
		"<@%s> Rappel activé pour familier %q sur %s\nProchain rappel: %s\n  ID: %s\n",
		m.Author.ID,
		pet,
		character,
		remind.NextRemind.In(h.timezone).Format(time.RFC1123),
		id.Hex(),
	)
	if _, err = h.discord.SendMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// RemoveRemind handles the remove command for the bot.
// Call it with `!remove <RemindID>`.
func (h *Handler) RemoveRemind(ctx context.Context, m *harmony.Message) {
	parts := strings.Split(m.Content, " ")
	if len(parts) != 2 {
		h.Help(ctx)

		return
	}

	id := parts[1]
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		log.Debug().Msg("ID invalid")

		message := fmt.Sprintf("<@%s> ID invalide.", m.Author.ID)
		if _, err := h.discord.SendMessage(ctx, message); err != nil {
			log.Error().Err(err).Msg("Unable to send message")

			return
		}

		return
	}

	logger := log.With().Str("id", id).Logger()

	remind, err := h.store.GetRemind(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to find remind")

		return
	}

	if remind.DiscordUserID != m.Author.ID {
		message := fmt.Sprintf("<@%s> Vous ne pouvez pas supprimer un rappel qui ne vous appartient pas.", m.Author.ID)
		if _, err = h.discord.SendMessage(ctx, message); err != nil {
			logger.Error().Err(err).Msg("Unable to send message")

			return
		}

		logger.Debug().Msg("Unable to disable reminder: wrong discordUserID")

		return
	}

	if err = h.store.RemoveRemind(ctx, id); err != nil {
		if errors.As(err, &store.NotFoundError{}) {
			logger.Debug().Err(err).Msg("Unable to find reminder")

			message := fmt.Sprintf("<@%s> Pas de rappel avec l'id: %q", m.Author.ID, id)
			if _, err = h.discord.SendMessage(ctx, message); err != nil {
				logger.Error().Err(err).Msg("Unable to send message")

				return
			}

			return
		}

		logger.Error().Err(err).Msg("Unable to remove reminder")

		return
	}

	h.reminder.SetUpdate()

	message := fmt.Sprintf("<@%s> Reminder %q supprimé", m.Author.ID, id)
	if _, err = h.discord.SendMessage(ctx, message); err != nil {
		logger.Error().Err(err).Msg("Unable to send message")

		return
	}
}

// Help handles all other commands.
func (h *Handler) Help(ctx context.Context) {
	message := "Commandes disponible:\n  - `!familiers`\n  - `!remind <Familier> <Personnage>`\n  - `!remove <ID>` "
	if _, err := h.discord.SendMessage(ctx, message); err != nil {
		log.Error().Err(err).Msg("Unable to send message")

		return
	}
}
