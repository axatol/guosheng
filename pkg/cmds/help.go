package cmds

import (
	"context"
	"fmt"
	"strings"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
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

func (cmd Help) OnMessageCommand(ctx context.Context, bot *discord.Bot, event *discordgo.MessageCreate, args []string) {
	var lines []string
	for _, c := range cmd.Commands {
		if msgCmd, ok := c.(discord.MessageCommandable); ok {
			line := fmt.Sprintf("`%s%s` - %s", bot.MessagePrefix, msgCmd.Name(), msgCmd.Description())
			lines = append(lines, line)
			continue
		}
	}

	if err := bot.SendMessageReply(ctx, event.Message, strings.Join(lines, "\n")); err != nil {
		log.Warn().Err(err).Send()
	}
}

func (cmd Help) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Help) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	var lines []string
	for _, c := range cmd.Commands {
		if msgCmd, ok := c.(discord.MessageCommandable); ok {
			line := fmt.Sprintf("`/%s` - %s", msgCmd.Name(), msgCmd.Description())
			lines = append(lines, line)
			continue
		}
	}

	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, strings.Join(lines, "\n")); err != nil {
		log.Warn().Err(err).Send()
	}
}
