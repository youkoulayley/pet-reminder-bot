package run

import "github.com/urfave/cli/v2"

// Command returns the run command.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run the ReminderBot",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "bot-token",
				Usage:    "Token for the bot",
				EnvVars:  []string{"BOT_TOKEN"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "channel-id",
				Usage:    "Channel where the bot will write message",
				EnvVars:  []string{"CHANNEL_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "mongo-uri",
				Usage:   "MongoDB connection string",
				EnvVars: []string{"MONGO_URI"},
				Value:   "mongodb://mongoadmin:secret@localhost:27017",
			},
		},
		Action: run,
	}
}
