package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony/discord"
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
		h.bot.Remind(ctx, m)
	case strings.HasPrefix(m.Content, "!remove"):
		h.bot.RemoveRemind(ctx, m)
	case strings.HasPrefix(m.Content, "!help"):
		h.bot.Help(ctx)
	default:
	}
}
