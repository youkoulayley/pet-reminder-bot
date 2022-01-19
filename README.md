# Pet Reminder Bot

This is a bot created for `Dofus Retro`, to remind me to feed my pets.

Available commands:
  - `!familiers`: list all pets available.
  - `!remind <PET_NAME> <CHARACTER_NAME>`: set a reminder for a pet on a specific character.
  - `!remove <ID>`: remove a reminder by its ID.

## How does this bot works?
If you enter this remind command, the bot will start a new reminder:
```
!remind Dragoune_Rose Dermatologue

@Youkoulayley Reminder activé pour familier "Dragoune_Rose" sur Dermatologue
Prochain rappel: Sun, 09 Jan 2022 16:00:35 UTC
ID: 61dac053b64a48a3de3520d3
```

This first reminder is set for the `current time + foodMinDuration` of the pet.

Each pet have a `foodMinDuration` and a `foodMaxduration`. If you feed your pet before the `foodMinDuration`, it will
not be happy, and same goes if you feed your pet after the `foodMaxDuration`.

The bot will send a message just after the `foodMinDuration` :
```
@Youkoulayley Il faut nourrir "Dragoune_Rose" sur Dermatologue
ID: 61dac053b64a65b3de7650d3
```

To notify the bot that you have fed your pet, just put a reaction on this message. Anything will do the trick.

If you don't, the bot will send you a message just after the `foodMaxDuration`:
```
@Youkoulayley "Dragoune_Rose" sur Dermatologue a râté 1 repas.
Prochain rappel: Sun, 09 Jan 2022 21:00:36 UTC
ID: 61dac053b64a65b3de7650d3
```

Once this message is sent the bot will set the reminder for the next `foodMinDuration` and a new cycle begins.

You can add a reaction to any message of the bot and the bot will start a new cycle for the current reminder.

## How to launch it?
You can use the docker compose to run the bot. It requires a mongo database for now (this is used to store the reminder
and allows the restart of the bot).
You just have to change :
  - The `BOT_TOKEN`: token of the bot on Discord.
  - The `BOT_CHANNEL_ID`: channel where the bot will write its messages.
  - The `BOT_TIMEZONE`: timezone for discord messages.

## Ideas
- Sent reminder by MP
- List reminders for a discord User
