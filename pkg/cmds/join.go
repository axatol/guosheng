package cmds

import (
	"context"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	// _ discord.MessageCommandable = (*Join)(nil)
	_ discord.ApplicationCommandable = (*Join)(nil)
)

type Join struct{}

func (cmd Join) Name() string {
	return "join"
}

func (cmd Join) Description() string {
	return "Join your voice channel"
}

func (cmd Join) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
	}
}

func (cmd Join) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	emoji := bot.GetEmojiForMessage("BatChest", "👋")
	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, emoji); err != nil {
		log.Warn().Err(err).Send()
	}

	if _, err := bot.JoinUserVoiceChannel(event.Member.User.ID); err != nil {
		log.Warn().Err(err).Send()
		return
	}
}
