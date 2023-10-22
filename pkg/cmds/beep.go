package cmds

import (
	"context"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.MessageHandler                       = (*Beep)(nil)
	_ discord.ApplicationCommandInteractionHandler = (*Beep)(nil)
)

type Beep struct{}

func (cmd Beep) Name() string {
	return "beep"
}

func (cmd Beep) Description() string {
	return "boop"
}

func (cmd Beep) OnMessage(ctx context.Context, bot *discord.Bot, event *discordgo.MessageCreate, args []string) {
	if err := bot.SendMessageReaction(ctx, event.Message, "🤖", "🤖"); err != nil {
		log.Warn().Err(err).Send()
	}

	if err := bot.SendMessageReply(ctx, event.Message, "boop"); err != nil {
		log.Warn().Err(err).Send()
	}
}

func (cmd Beep) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Beep) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "boop"); err != nil {
		log.Warn().Err(err).Send()
	}
}
