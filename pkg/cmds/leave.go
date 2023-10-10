package cmds

import (
	"context"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	// _ discord.MessageCommandable = (*Leave)(nil)
	_ discord.ApplicationCommandable = (*Leave)(nil)
)

type Leave struct{}

func (cmd Leave) Name() string {
	return "leave"
}

func (cmd Leave) Description() string {
	return "Leave your voice channel"
}

func (cmd Leave) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Leave) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	emoji := bot.GetEmojiForMessage("FeelsCarlosMan", "👋")
	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, emoji); err != nil {
		log.Warn().Err(err).Send()
	}

	if err := bot.LeaveUserVoiceChannel(event.Member.User.ID); err != nil {
		log.Warn().Err(err).Send()
		return
	}
}
