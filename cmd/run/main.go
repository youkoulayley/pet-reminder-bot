package run

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skwair/harmony"
	"github.com/urfave/cli/v2"
	"github.com/youkoulayley/pet-reminder-bot/pkg/handlers"
	"github.com/youkoulayley/pet-reminder-bot/pkg/reminder"
	"github.com/youkoulayley/pet-reminder-bot/pkg/store"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const databaseName = "pet-reminder-bot"

func run(ctx *cli.Context) error {
	discordClient, err := harmony.NewClient(ctx.String("bot-token"))
	if err != nil {
		return fmt.Errorf("create discord client: %w", err)
	}

	bot, err := discordClient.User(ctx.Context, "@me")
	if err != nil {
		return fmt.Errorf("get bot user: %w", err)
	}

	channel := discordClient.Channel(ctx.String("channel-id"))

	opts := options.Client().
		ApplyURI(ctx.String("mongo-uri")).
		SetSocketTimeout(2 * time.Second).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.NewClient(opts)
	if err != nil {
		return fmt.Errorf("create MongoDB client: %w", err)
	}

	if err = client.Connect(ctx.Context); err != nil {
		return fmt.Errorf("connect db: %w", err)
	}

	defer func() { _ = client.Disconnect(ctx.Context) }()

	s, err := store.New(client, databaseName)
	if err != nil {
		return fmt.Errorf("new store: %w", err)
	}

	r, err := reminder.New(&s, channel)
	if err != nil {
		return fmt.Errorf("new reminder: %w", err)
	}

	go r.Run(ctx.Context)

	if err = s.Bootstrap(ctx.Context); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}

	tz, err := time.LoadLocation(ctx.String(flagBotTimezone))
	if err != nil {
		return fmt.Errorf("load location %q: %w", flagBotTimezone, err)
	}

	h := handlers.NewHandler(bot, channel, &s, r, tz)

	discordClient.OnMessageCreate(h.MessageCreate)
	discordClient.OnMessageReactionAdd(h.ReactionAdd)

	if err = discordClient.Connect(ctx.Context); err != nil {
		return fmt.Errorf("discord client connect: %w", err)
	}

	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	return nil
}
