package cmds

import (
	"context"
	"fmt"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
)

var (
	_ discord.Interactable = (*Beep)(nil)
	_ discord.Commandable  = (*Beep)(nil)
)

type Beep struct{}

func (cmd Beep) Name() string {
	return "beep"
}

func (cmd Beep) Description() string {
	return "boop"
}

func (cmd Beep) OnMessage(ctx context.Context, bot *discord.Bot, event *discordgo.MessageCreate, args []string) error {
	if err := bot.Session.MessageReactionAdd(event.ChannelID, event.Message.ID, "🤖"); err != nil {
		return fmt.Errorf("failed to react to message: %s", err)
	}

	if _, err := bot.Session.ChannelMessageSendReply(event.ChannelID, "boop", event.Message.Reference(), discordgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("failed to reply to message: %s", err)
	}

	return nil
}

func (cmd Beep) Interaction() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Beep) OnApplicationCommandInteraction(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) error {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "boop"},
	}

	if err := bot.Session.InteractionRespond(event.Interaction, &response, discordgo.WithContext(ctx)); err != nil {
		return fmt.Errorf("failed to respond to interaction: %s", err)
	}

	return nil
}
