package cmds

import (
	"context"
	"fmt"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.MessageCommandable     = (*Beep)(nil)
	_ discord.ApplicationCommandable = (*Beep)(nil)
)

type Beep struct{}

func (cmd Beep) Name() string {
	return "beep"
}

func (cmd Beep) Description() string {
	return "boop"
}

func (cmd Beep) OnMessageCommand(ctx context.Context, bot *discord.Bot, event *discordgo.MessageCreate, args []string) {
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
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "boop"},
	}

	if err := bot.Session.InteractionRespond(event.Interaction, &response, discord.WithRequestOptions(ctx)...); err != nil {
		log.Warn().Err(fmt.Errorf("failed to respond to interaction: %s", err)).Send()
	}
}
