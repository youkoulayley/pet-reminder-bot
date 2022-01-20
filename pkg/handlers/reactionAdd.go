package handlers

import (
	"context"
	"time"

	"github.com/skwair/harmony"
)

// ReactionAdd gets all reactions created.
func (h Handler) ReactionAdd(m *harmony.MessageReaction) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h.bot.NewCycle(ctx, m)
}
