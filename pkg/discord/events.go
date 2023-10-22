package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (b *Bot) onEvent(session *discordgo.Session, event *discordgo.Event) {
	log.Trace().Str("event", event.Type).
		Any("data", event.Struct).
		Send()
}

func (b *Bot) onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	var user *discordgo.User
	if event.Member != nil && event.Member.User != nil {
		user = event.Member.User
	} else if event.User != nil {
		user = event.User
	}

	if user.Bot {
		return
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

	switch event.Type {
	case discordgo.InteractionApplicationCommand:
		b.onInteractionApplicationCommand(event, log)
		return

	case discordgo.InteractionMessageComponent:
		b.onInteractionMessageComponent(event, log)
		return

	default:
		log.Warn().
			Any("data", event.Data).
			Msg("unhandled interaction type")
		return
	}
}

func (b *Bot) onInteractionApplicationCommand(event *discordgo.InteractionCreate, log zerolog.Logger) {
	data := event.ApplicationCommandData()
	log = log.With().
		Any("application_command_data", data).
		Str("command_name", data.Name).
		Any("command_options", data.Options).
		Logger()

	cmd, ok := b.Commands[data.Name]
	if !ok {
		log.Debug().
			Err(fmt.Errorf("invalid command: %s", ErrCommandNotImplemented)).
			Send()
		return
	}

	handler, ok := cmd.(ApplicationCommandInteractionHandler)
	if !ok {
		log.Error().
			Err(fmt.Errorf("invalid command: %s", ErrNotApplicationCommandInteractionHandler)).
			Send()
		return
	}

	log.Info().Send()
	go handler.OnApplicationCommand(context.Background(), b, event, &data)
}

func (b *Bot) onInteractionMessageComponent(event *discordgo.InteractionCreate, log zerolog.Logger) {
	data := event.MessageComponentData()
	log = log.With().
		Any("message_component_data", data).
		Str("custom_id", data.CustomID).
		Any("values", data.Values).
		Logger()

	idParts := strings.Split(data.CustomID, ":")
	if len(idParts) != 2 {
		log.Debug().
			Err(fmt.Errorf("invalid custom id format")).
			Send()
		return
	}

	commandName := idParts[0]
	interactionID := idParts[1]
	log = log.With().
		Str("command_name", commandName).
		Str("interaction_id", interactionID).
		Logger()

	cmd, ok := b.Commands[commandName]
	if !ok {
		log.Debug().
			Err(fmt.Errorf("invalid command: %s", ErrCommandNotImplemented)).
			Send()
		return
	}

	handler, ok := cmd.(MessageComponentInteractionHandler)
	if !ok {
		log.Error().
			Err(fmt.Errorf("invalid command: %s", ErrNotMessageComponentInteractionHandler)).
			Send()
		return
	}

	log.Info().Send()
	go handler.OnMessageComponent(context.Background(), b, event, &data)
}

func (b *Bot) onMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	if event.Author.Bot {
		return
	}

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

	log.Info().
		Str("event", "MESSAGE_CREATE").
		Str("message_id", event.ID).
		Bool("user_is_bot", event.Author.Bot).
		Send()

	cmd, ok := b.Commands[name]
	if !ok {
		log.Warn().
			Err(fmt.Errorf("invalid command: %s", ErrCommandNotImplemented)).
			Send()
		return
	}

	command, ok := cmd.(MessageHandler)
	if !ok {
		log.Debug().
			Err(fmt.Errorf("invalid command: %s", ErrNotMessageHandler)).
			Send()
		return
	}

	go command.OnMessage(context.Background(), b, event, args)
}

func (b *Bot) onMessageReactionAdd(session *discordgo.Session, event *discordgo.MessageReactionAdd) {
	if event.Member.User.Bot {
		return
	}

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
		Dur("retry_after", event.RetryAfter).
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

	if err := session.UpdateWatchStatus(0, "Scooby Doo"); err != nil {
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
