package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

// ReactionAdd gets all reactions created.
func (h Handler) ReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	message, err := s.ChannelMessage(h.channelID, m.MessageID)
	if err != nil {
		log.Error().Err(err).Msg("Unable to find message")
		return
	}

	msg := strings.Split(message.Content, "ID:")
	id := strings.TrimSpace(msg[1])
	if id == "" {
		log.Error().Msg("ID cannot be empty")
		return
	}

	ctx := context.Background()
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
