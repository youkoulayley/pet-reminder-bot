package run

import (
	"github.com/ettle/strcase"
	"github.com/urfave/cli/v2"
)

const (
	flagLogLevel     = "log-level"
	flagBotToken     = "bot-token"
	flagBotChannelID = "bot-channel-id"
	flagBotTimezone  = "bot-timezone"
	flagMongoURI     = "mongo-uri"
)

// Command returns the run command.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run the PetReminderBot",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    flagLogLevel,
				Usage:   "Log level",
				EnvVars: []string{strcase.ToSNAKE(flagLogLevel)},
				Value:   "debug",
			},
			&cli.StringFlag{
				Name:     flagBotToken,
				Usage:    "Token for the bot",
				EnvVars:  []string{strcase.ToSNAKE(flagBotToken)},
				Required: true,
			},
			&cli.StringFlag{
				Name:     flagBotChannelID,
				Usage:    "Channel where the bot will write message",
				EnvVars:  []string{strcase.ToSNAKE(flagBotChannelID)},
				Required: true,
			},
			&cli.StringFlag{
				Name:    flagBotTimezone,
				Usage:   "Print the message with this timezone",
				EnvVars: []string{strcase.ToSNAKE(flagBotTimezone)},
				Value:   "Europe/Paris",
			},
			&cli.StringFlag{
				Name:    flagMongoURI,
				Usage:   "MongoDB connection string",
				EnvVars: []string{strcase.ToSNAKE(flagMongoURI)},
				Value:   "mongodb://mongoadmin:secret@localhost:27017",
			},
		},
		Action: run,
	}
}
