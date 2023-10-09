package cmds

import (
	"context"
	"fmt"
	"strings"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

var (
	_ discord.MessageCommandable     = (*Help)(nil)
	_ discord.ApplicationCommandable = (*Help)(nil)
)

type Help struct{ Commands map[string]any }

func (cmd Help) Name() string {
	return "help"
}

func (cmd Help) Description() string {
	return "display all available commands"
}

func (cmd Help) OnMessageCommand(ctx context.Context, bot *discord.Bot, event *discordgo.MessageCreate, args []string) error {
	var lines []string
	for _, c := range cmd.Commands {
		if msgCmd, ok := c.(discord.MessageCommandable); ok {
			line := fmt.Sprintf("`%s%s` - %s", bot.MessagePrefix, msgCmd.Name(), msgCmd.Description())
			lines = append(lines, line)
			continue
		}
	}

	if _, err := bot.Session.ChannelMessageSendReply(event.ChannelID, strings.Join(lines, "\n"), event.Message.Reference(), discord.WithRequestOptions(ctx)...); err != nil {
		return fmt.Errorf("failed to reply to message: %s", err)
	}

	return nil
}

func (cmd Help) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Help) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) error {
	var lines []string
	for _, c := range cmd.Commands {
		if msgCmd, ok := c.(discord.MessageCommandable); ok {
			line := fmt.Sprintf("`/%s` - %s", msgCmd.Name(), msgCmd.Description())
			lines = append(lines, line)
			continue
		}
	}

	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: strings.Join(lines, "\n")},
	}

	if err := bot.Session.InteractionRespond(event.Interaction, &response, discord.WithRequestOptions(ctx)...); err != nil {
		return fmt.Errorf("failed to respond to interaction: %s", err)
	}

	return nil
}
