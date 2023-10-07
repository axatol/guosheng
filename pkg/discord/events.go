package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

func (b *Bot) onGuildCreate(session *discordgo.Session, event *discordgo.GuildCreate) {
	log.Info().
		Str("event", "GUILD_CREATE").
		Str("guild_id", event.ID).
		Str("guild_name", event.Name).
		Send()

	b.Guilds[event.ID] = Guild{event.Guild}
}

func (b *Bot) onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	var user *discordgo.User
	if event.Member != nil && event.Member.User != nil {
		user = event.Member.User
	} else if event.User != nil {
		user = event.User
	}

	log := log.With().
		Str("user_username", user.Username).
		Str("channel_id", event.ChannelID).
		Str("guild_id", event.GuildID).
		Str("interaction_type", event.Data.Type().String()).
		Logger()

	log.Info().
		Str("event", "INTERACTION_CREATE").
		Str("interaction_id", event.ID).
		Bool("user_is_bot", user.Bot).
		Send()

	switch data := (event.Data).(type) {
	case discordgo.ApplicationCommandInteractionData:
		log = log.With().
			Any("interaction_data_application_command", data).
			Str("command_name", data.Name).
			Any("command_options", data.Options).
			Logger()

		interaction, ok := b.interactions[data.Name]
		if !ok {
			return
		}

		interactible, ok := interaction.(ApplicationCommandInteractive)
		if !ok {
			log.Warn().Msg("matched interaction data cannot handle data type")
		}

		if err := interactible.OnApplicationCommandInteraction(context.Background(), b, event, &data); err != nil {
			log.Error().Err(err).Send()
		}

	case discordgo.MessageComponentInteractionData:
		log.Warn().
			Any("interaction_data_message_component", data).
			Msg("unsupported interaction type")
		return

	case discordgo.ModalSubmitInteractionData:
		log.Warn().
			Any("interaction_data_modal_submit", data).
			Msg("unsupported interaction type")
		return
	}
}

func (b *Bot) onMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	content, ok := strings.CutPrefix(event.Content, b.MessagePrefix)
	if !ok {
		return
	}

	parts := strings.Fields(content)
	name := parts[0]
	args := parts[1:]

	log := log.With().
		Str("user_username", event.Author.Username).
		Str("channel_id", event.ChannelID).
		Str("guild_id", event.GuildID).
		Str("command_name", name).
		Strs("command_args", args).
		Logger()

	cmd, ok := b.commands[name]
	if !ok {
		log = log.With().Str("msg", "command not found").Logger()
		return
	}

	log.Info().
		Str("event", "MESSAGE_CREATE").
		Str("message_id", event.ID).
		Bool("user_is_bot", event.Author.Bot).
		Send()

	if event.Author.Bot {
		return
	}

	if err := cmd.OnMessage(context.Background(), b, event, args); err != nil {
		log.Error().
			Err(fmt.Errorf("invocation failed: %s", err)).
			Send()
	}
}

func (b *Bot) onMessageReactionAdd(session *discordgo.Session, event *discordgo.MessageReactionAdd) {
	log.Info().
		Str("event", "MESSAGE_REACTION_ADD").
		Str("user_username", event.Member.User.Username).
		Bool("user_is_bot", event.Member.User.Bot).
		Str("guild_id", event.GuildID).
		Str("message_id", event.MessageID).
		Str("emoji_id", event.Emoji.ID).
		Str("emoji_name", event.Emoji.Name).
		Send()
}

func (b *Bot) onRateLimit(session *discordgo.Session, event *discordgo.RateLimit) {
	log.Info().
		Str("event", "RATE_LIMIT").
		Any("payload", event).
		Send()
}

func (b *Bot) onReady(session *discordgo.Session, event *discordgo.Ready) {
	log.Info().
		Str("event", "READY").
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

func (b *Bot) onVoiceServerUpdate(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
	log.Info().
		Str("event", "VOICE_SERVER_UPDATE").
		Str("endpoint", event.Endpoint).
		Str("guild_id", event.GuildID).
		Send()
}

func (b *Bot) onVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
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

	log.Info().
		Str("event", "VOICE_STATE_UPDATE").
		Str("user_username", event.Member.User.Username).
		Str("user_action", action).
		Bool("user_is_bot", event.Member.User.Bot).
		Str("channel_id_new", event.ChannelID).
		Str("channel_id_old", oldChannelId).
		Str("guild_id", event.GuildID).
		Send()
}
