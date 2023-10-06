package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func onConnect(session *discordgo.Session, event *discordgo.Connect) {
	log.Info().Str("event", "Connect").Send()
}

func onDisconnect(session *discordgo.Session, event *discordgo.Disconnect) {
	log.Info().Str("event", "Disconnect").Send()
}

func onEvent(session *discordgo.Session, event *discordgo.Event) {
	filter := []string{
		"CONNECT",
		"DISCONNECT",
		"GUILD_CREATE",
		"INTERACTION_CREATE",
		"MESSAGE_CREATE",
		"MESSAGE_REACTION_ADD",
		"RATE_LIMIT",
		"READY",
		"VOICE_SERVER_UPDATE",
		"VOICE_STATE_UPDATE",
	}

	for _, f := range filter {
		if event.Type == f {
			return
		}
	}

	log.Debug().
		Str("event", "Event").
		Any("payload", event).
		Send()
}

func onGuildCreate(bot Bot) any {
	return func(session *discordgo.Session, event *discordgo.GuildCreate) {
		log.Info().
			Str("event", "GuildCreate").
			Str("guild_id", event.Guild.ID).
			Str("guild_name", event.Guild.Name).
			Send()

		bot.guilds[event.ID] = Guild{event.Guild}
	}
}

func onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	log.Info().
		Str("event", "InteractionCreate").
		Any("payload", event).
		Send()
}

func onMessageCreate(discordBotPrefix string) any {
	return func(session *discordgo.Session, event *discordgo.MessageCreate) {
		content, _ := strings.CutPrefix(event.Content, discordBotPrefix)

		log.Info().
			Str("event", "MessageCreate").
			Str("user_id", event.Author.ID).
			Str("user_username", event.Author.Username).
			Bool("user_is_bot", event.Author.Bot).
			Str("channel_id", event.ChannelID).
			Str("message_id", event.ID).
			Str("content", content).
			Send()
	}
}

func onMessageReactionAdd(session *discordgo.Session, event *discordgo.MessageReactionAdd) {
	log.Info().
		Str("event", "MessageReactionAdd").
		Str("user_id", event.Member.User.ID).
		Str("user_username", event.Member.User.Username).
		Bool("user_is_bot", event.Member.User.Bot).
		Str("guild_id", event.GuildID).
		Str("message_id", event.MessageID).
		Send()
}

func onRateLimit(session *discordgo.Session, event *discordgo.RateLimit) {
	log.Info().
		Str("event", "RateLimit").
		Any("payload", event).
		Send()
}

func onReady(session *discordgo.Session, event *discordgo.Ready) {
	log.Info().
		Str("event", "Ready").
		Str("username", event.User.Username).
		Str("session_id", event.SessionID).
		Str("application_id", event.Application.ID).
		Int("guild_count", len(event.Guilds)).
		Int("version", event.Version).
		Send()

	usd := discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{{Details: "Plying the dildont"}},
	}

	if err := session.UpdateStatusComplex(usd); err != nil {
		log.Error().Err(fmt.Errorf("failed to update bot status: %s", err)).Send()
	}
}

func onVoiceServerUpdate(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
	log.Info().
		Str("event", "VoiceServerUpdate").
		Any("payload", event).
		Send()
}

func onVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
	oldChannelId := ""
	if event.BeforeUpdate != nil {
		oldChannelId = event.BeforeUpdate.ChannelID
	}

	action := ""
	if oldChannelId == "" && event.ChannelID != "" {
		action = "joined"
	} else if oldChannelId != "" && event.ChannelID == "" {
		action = "left"
	}

	if action == "" {
		return
	}

	log := log.Info().
		Str("event", "VoiceStateUpdate").
		Str("user_id", event.Member.User.ID).
		Str("user_username", event.Member.User.Username).
		Str("user_action", action).
		Bool("user_is_bot", event.Member.User.Bot).
		Str("channel_id_new", event.ChannelID).
		Str("channel_id_old", oldChannelId).
		Str("guild_id", event.GuildID)

	if event.BeforeUpdate != nil {
		log = log.Str("channel_id_before", event.BeforeUpdate.ChannelID)
	}

	log.Send()
}
