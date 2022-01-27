package handlers

import (
	"context"
	"time"

	"github.com/skwair/harmony"
	"github.com/youkoulayley/pet-reminder-bot/pkg/bot"
)

// ReactionAdd gets all reactions created.
func (h Handler) ReactionAdd(m *harmony.MessageReaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := bot.NewCycleConfig{AuthorID: m.UserID, MessageID: m.MessageID}
	h.bot.NewCycle(ctx, cfg)
}
