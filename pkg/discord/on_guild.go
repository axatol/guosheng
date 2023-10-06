package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func onGuildCreate(bot Bot) any {
	return func(session *discordgo.Session, event *discordgo.GuildCreate) {
		log.Debug().
			Str("event", "GuildCreate").
			Str("guild_id", event.Guild.ID).
			Str("guild_name", event.Guild.Name).
			// Any("payload", event).
			Send()

		bot.guilds[event.ID] = Guild{event.Guild}
	}
}
