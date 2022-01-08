package run

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/youkoulayley/reminderbot/pkg/handlers"
	"github.com/youkoulayley/reminderbot/pkg/reminder"
	"github.com/youkoulayley/reminderbot/pkg/store"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func run(ctx *cli.Context) error {
	dg, err := discordgo.New("Bot " + ctx.String("bot-token"))
	if err != nil {
		return errors.New("unable to start discord session")
	}

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

	s, err := store.New(client)
	if err != nil {
		return fmt.Errorf("new store: %w", err)
	}

	r, err := reminder.New(&s, dg, ctx.String("channel-id"))
	if err != nil {
		return fmt.Errorf("new reminder: %w", err)
	}
	go r.Run(ctx.Context)

	if err = s.Bootstrap(ctx.Context); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}

	h := handlers.NewHandler(&s, r, ctx.String("channel-id"))

	dg.AddHandler(h.MessageCreate)
	dg.AddHandler(h.ReactionAdd)

	dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = dg.Open()
	if err != nil {
		return err
	}

	log.Info().Msg("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err = dg.Close(); err != nil {
		return err
	}

	return nil
}
