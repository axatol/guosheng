package cmds

import (
	"context"
	"fmt"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

var (
	_ discord.MessageCommandable     = (*Shutdown)(nil)
	_ discord.ApplicationCommandable = (*Shutdown)(nil)
)

type Shutdown struct{ Shutdown func() }

func (cmd Shutdown) Name() string {
	return "shutdown"
}

func (cmd Shutdown) Description() string {
	return "shutdown the bot"
}

func (cmd Shutdown) OnMessageCommand(ctx context.Context, bot *discord.Bot, event *discordgo.MessageCreate, args []string) (err error) {
	if emoji, ok := bot.Emojis["FeelsCarlosMan"]; ok {
		if err = bot.Session.MessageReactionAdd(event.ChannelID, event.ID, emoji.ID, discord.WithRequestOptions(ctx)...); err != nil {
			err = fmt.Errorf("failed to react to message: %s", err)
		}
	}

	cmd.Shutdown()
	return err
}

func (cmd Shutdown) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Shutdown) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) (err error) {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "😶‍🌫️"},
	}

	if emoji, ok := bot.Emojis["FeelsCarlosMan"]; ok {
		response.Data = &discordgo.InteractionResponseData{Content: emoji.MessageFormat()}
	}

	if err = bot.Session.InteractionRespond(event.Interaction, &response, discord.WithRequestOptions(ctx)...); err != nil {
		err = fmt.Errorf("failed to respond to interaction: %s", err)
	}

	cmd.Shutdown()
	return err
}
