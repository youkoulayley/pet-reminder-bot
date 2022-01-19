package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReactionAdd gets all reactions created.
func (h Handler) ReactionAdd(m *harmony.MessageReaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message, err := h.discord.Message(ctx, m.MessageID)
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

	remind, err := h.store.GetRemind(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Unable to get remind")

		return
	}

	if remind.DiscordUserID != m.UserID {
		log.Error().Msg("Invalid user")

		return
	}

	pet, err := h.store.GetPet(ctx, remind.PetName)
	if err != nil {
		log.Error().Err(err).Msg("Invalid pet")

		return
	}

	remind.MissedReminder = 0
	remind.ReminderSent = false
	remind.NextRemind = time.Now().Add(pet.FoodMinDuration)
	remind.TimeoutRemind = time.Now().Add(pet.FoodMaxDuration)

	if err = h.store.UpdateRemind(ctx, remind); err != nil {
		log.Error().Err(err).Msg("Unable to update remind")

		return
	}

	h.reminder.SetUpdate()
}
