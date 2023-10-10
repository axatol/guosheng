package cmds

import (
	"context"
	"fmt"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	// _ discord.MessageCommandable     = (*Shutdown)(nil)
	_ discord.ApplicationCommandable = (*Shutdown)(nil)
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
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: emoji},
	}

	if err := bot.Session.InteractionRespond(event.Interaction, &response, discord.WithRequestOptions(ctx)...); err != nil {
		log.Warn().Err(fmt.Errorf("failed to respond to interaction: %s", err)).Send()
	}
}
