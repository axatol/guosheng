package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func onMessageCreate(discordBotPrefix string) any {
	return func(session *discordgo.Session, event *discordgo.MessageCreate) {
		content, _ := strings.CutPrefix(event.Content, discordBotPrefix)

		log.Debug().
			Str("event", "MessageCreate").
			Str("user_id", event.Author.ID).
			Str("user_username", event.Author.Username).
			Bool("user_is_bot", event.Author.Bot).
			// Strs("user_roles", event.Member.Roles).
			Str("channel_id", event.ChannelID).
			Str("guild_id", event.GuildID).
			Str("message_id", event.ID).
			Str("content", content).
			// Any("payload", event).
			Send()
	}
}

func onMessageReactionAdd(session *discordgo.Session, event *discordgo.MessageReactionAdd) {
	log.Debug().
		Str("event", "MessageReactionAdd").
		Str("user_id", event.Member.User.ID).
		Str("user_username", event.Member.User.Username).
		Bool("user_is_bot", event.Member.User.Bot).
		// Strs("user_roles", event.Member.Roles).
		Str("channel_id", event.ChannelID).
		Str("guild_id", event.GuildID).
		Str("message_id", event.MessageID).
		// Any("payload", event).
		Send()
}
