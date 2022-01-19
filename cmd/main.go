package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/youkoulayley/pet-reminder-bot/cmd/run"
)

func main() {
	app := &cli.App{
		Name: "Pet Reminder Bot",
		Commands: []*cli.Command{
			run.Command(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Error during execution")

		return
	}
}
