package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony/discord"
	"github.com/youkoulayley/pet-reminder-bot/pkg/bot"
)

// MessageCreate gets all message created.
// All messages send by the bot are ignored.
func (h *Handler) MessageCreate(m *discord.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if h.botUser.ID == m.Author.ID {
		log.Debug().Msg("Skipping message sent by me")

		return
	}

	switch {
	case strings.HasPrefix(m.Content, "!familiers"):
		h.bot.ListPets(ctx)
	case strings.HasPrefix(m.Content, "!remind"):
		cfg, err := h.handleRemindConfig(m)
		if err != nil {
			h.bot.Help(ctx)

			return
		}

		h.bot.Remind(ctx, cfg)
	case strings.HasPrefix(m.Content, "!remove"):
		cfg, err := h.handleRemoveRemindConfig(m)
		if err != nil {
			h.bot.Help(ctx)

			return
		}

		h.bot.RemoveRemind(ctx, cfg)
	case strings.HasPrefix(m.Content, "!help"):
		h.bot.Help(ctx)
	default:
	}
}

func (h *Handler) handleRemindConfig(m *discord.Message) (bot.RemindConfig, error) {
	parts := strings.Split(m.Content, " ")
	if len(parts) != 3 {
		return bot.RemindConfig{}, errors.New("command invalid")
	}

	pet := parts[1]
	character := parts[2]

	if pet == "" || character == "" {
		return bot.RemindConfig{}, errors.New("pet or character missing")
	}

	return bot.RemindConfig{
		AuthorID:  m.Author.ID,
		Pet:       pet,
		Character: character,
	}, nil
}

func (h *Handler) handleRemoveRemindConfig(m *discord.Message) (bot.RemoveRemindConfig, error) {
	parts := strings.Split(m.Content, " ")
	if len(parts) != 2 {
		return bot.RemoveRemindConfig{}, errors.New("command invalid")
	}

	id := parts[1]
	if id == "" {
		return bot.RemoveRemindConfig{}, errors.New("id is missing")
	}

	return bot.RemoveRemindConfig{
		AuthorID: m.Author.ID,
		ID:       id,
	}, nil
}
