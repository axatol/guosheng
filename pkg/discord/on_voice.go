package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func onVoiceServerUpdate(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
	log.Debug().Str("event", "VoiceServerUpdate").Any("payload", event).Send()
}

func onVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	log.Debug().
		Str("event", "VoiceStateUpdate").
		Str("user_id", event.Member.User.ID).
		Str("user_username", event.Member.User.Username).
		Bool("user_is_bot", event.Member.User.Bot).
		Bool("user_left", event.ChannelID == "").
		// Strs("user_roles", event.Member.Roles).
		Str("channel_id", event.ChannelID).
		Str("guild_id", event.GuildID).
		// Any("payload", event).
		Send()
}
