package cmds

import (
	"context"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.ApplicationCommandInteractionHandler = (*Shutdown)(nil)
)

type Shutdown struct{ Shutdown func() }

func (cmd Shutdown) Name() string {
	return "shutdown"
}

func (cmd Shutdown) Description() string {
	return "shutdown the bot"
}

func (cmd Shutdown) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Shutdown) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	defer cmd.Shutdown()

	emoji := bot.GetEmojiForMessage("FeelsCarlosMan", "😶‍🌫️")
	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, emoji); err != nil {
		log.Warn().Err(err).Send()
	}
}
