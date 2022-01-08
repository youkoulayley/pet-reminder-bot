package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/youkoulayley/reminderbot/pkg/store"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageCreate gets all message created.
// All messages send by the bot are ignored.
func (h Handler) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		log.Debug().Msg("Skipping message send by me")
		return
	}

	if m.Content == "!familiers" {
		h.ListPets(s, m)
	}

	if strings.HasPrefix(m.Content, "!remind") {
		h.Remind(s, m)
	}

	if strings.HasPrefix(m.Content, "!remove") {
		h.RemoveRemind(s, m)
	}
}

// ListPets handles the familiers command for the bot.
// Call it with `!familiers`.
func (h Handler) ListPets(s *discordgo.Session, m *discordgo.MessageCreate) {
	pets, err := h.store.ListPets(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Unable to list pets")
		return
	}

	if _, err := s.ChannelMessageSend(h.channelID, pets.String()); err != nil {
		log.Error().Err(err).Msg("Unable to send message")
		return
	}
}

// Remind handles the remind command for the bot.
// Call it with `!remind <PetName> <CharacterName>`
// PetName can be found with the ListPets command.
func (h Handler) Remind(s *discordgo.Session, m *discordgo.MessageCreate) {
	splitted := strings.Split(m.Content, " ")

	pet := splitted[1]
	character := splitted[2]

	ctx := context.Background()
	petDuration, err := h.store.GetPet(ctx, pet)
	if err != nil {
		if _, err = s.ChannelMessageSend(
			h.channelID,
			fmt.Sprintf("%q n'existe pas. `!familiers` pour connaître la liste des familiers gérés.", pet)); err != nil {
			log.Error().Err(err).Msg("Unable to send message")
			return
		}
		return
	}
	nextReminder := time.Now().Add(petDuration.FoodMinDuration)

	id := primitive.NewObjectID()
	if err = h.store.CreateRemind(ctx, store.Remind{
		ID:            id,
		DiscordUserID: m.Author.ID,
		PetName:       pet,
		Character:     character,
		NextRemind:    nextReminder,
		TimeoutRemind: time.Now().Add(petDuration.FoodMaxDuration),
	}); err != nil {
		return
	}

	if _, err = s.ChannelMessageSend(
		h.channelID,
		fmt.Sprintf("<@%s> Reminder activé pour familier %q sur %s\nProchain rappel: %s\nID: %s", m.Author.ID, pet, character, nextReminder.Format(time.RFC1123), id.Hex())); err != nil {
		log.Error().Err(err).Msg("Unable to send message")
		return
	}

	h.reminder.SetUpdate()
}

// RemoveRemind handles the remove command for the bot.
// Call it with `!remove <RemindID>`.
func (h Handler) RemoveRemind(s *discordgo.Session, m *discordgo.MessageCreate) {
	splitted := strings.Split(m.Content, " ")

	id := splitted[1]

	ctx := context.Background()
	remind, err := h.store.GetRemind(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("Unable to find remind")
		return
	}

	if remind.DiscordUserID != m.Author.ID {
		log.Debug().Msg("Unable to disable reminder: wrong discordUserID")
		return
	}

	if err = h.store.RemoveRemind(ctx, id); err != nil {
		if _, err = s.ChannelMessageSend(
			h.channelID,
			fmt.Sprintf("Pas de reminder avec l'id: %q", id)); err != nil {
			log.Error().Err(err).Msg("Unable to send message")
			return
		}
		return
	}

	if _, err = s.ChannelMessageSend(
		h.channelID,
		fmt.Sprintf("<@%s> Reminder %q supprimé", m.Author.ID, id)); err != nil {
		log.Error().Err(err).Msg("Unable to send message")
		return
	}

	h.reminder.SetUpdate()
}
