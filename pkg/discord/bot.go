package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/axatol/guosheng/pkg/config"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

const (
	discordIntents = discordgo.IntentGuilds |
		discordgo.IntentGuildVoiceStates |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildMessageReactions |
		discordgo.IntentDirectMessages |
		discordgo.IntentDirectMessageReactions |
		discordgo.IntentMessageContent
)

type BotOptions struct {
	AppID         string
	BotToken      string
	MessagePrefix string
}

type Bot struct {
	BotOptions
	Session  *discordgo.Session
	Commands map[string]any
}

func NewBot(opts BotOptions) (*Bot, error) {
	if opts.AppID == "" {
		return nil, fmt.Errorf("discord app id is required")
	}

	if opts.BotToken == "" {
		return nil, fmt.Errorf("discord bot token is required")
	}

	if opts.MessagePrefix == "" {
		return nil, fmt.Errorf("discord message prefix is required")
	}

	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.DiscordBotToken))
	if err != nil {
		return nil, fmt.Errorf("failed to start discord session: %s", err)
	}

	bot := Bot{
		BotOptions: opts,
		Session:    session,
		Commands:   make(map[string]any),
	}

	bot.Session.Identify.Intents = discordIntents
	bot.Session.AddHandler(bot.onEvent)
	bot.Session.AddHandler(bot.onInteractionCreate)
	bot.Session.AddHandler(bot.onMessageCreate)
	bot.Session.AddHandler(bot.onMessageReactionAdd)
	bot.Session.AddHandler(bot.onRateLimit)
	bot.Session.AddHandler(bot.onReady)
	bot.Session.AddHandler(bot.onVoiceServerUpdate)
	bot.Session.AddHandler(bot.onVoiceStateUpdate)

	return &bot, nil
}

func (b *Bot) Open(ctx context.Context, deadline time.Duration) error {
	log.Debug().Msg("opening connection to discord")

	wait := make(chan struct{}, 1)
	b.Session.AddHandlerOnce(func(s *discordgo.Session, e *discordgo.Ready) {
		wait <- struct{}{}
	})

	if err := b.Session.Open(); err != nil {
		return fmt.Errorf("failed to open websocket connection to discord: %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to connect to discord: %s", ctx.Err())
	case <-wait:
		return nil
	}
}

func (b *Bot) Close() error {
	// TODO close voice connections?
	for _, vc := range b.Session.VoiceConnections {
		if err := vc.Disconnect(); err != nil {
			log.Error().Err(fmt.Errorf("failed to disconnect voice channel %s: %s", vc.ChannelID, err)).Send()
		}
	}

	return b.Session.Close()
}

func (b *Bot) Ready(ctx context.Context) bool {
	return b.Session.DataReady
}

func (b *Bot) Health(ctx context.Context) (any, error) {
	metadata := map[string]any{
		"data_websocket_ready": b.Session.DataReady,
		"heartbeat_latency_ms": b.Session.HeartbeatLatency().Milliseconds(),
	}

	voiceConnections := map[string]bool{}
	for _, vc := range b.Session.VoiceConnections {
		voiceConnections[vc.GuildID+":"+vc.ChannelID] = vc.Ready
	}

	metadata["voice_connections"] = voiceConnections

	if !b.Session.DataReady {
		return metadata, fmt.Errorf("data websocket not ready")
	}

	return metadata, nil
}

func (b *Bot) RegisterInteractions(ctx context.Context) error {
	existing, err := b.Session.ApplicationCommands(b.AppID, "", RequestOptions(ctx))
	if err != nil {
		return fmt.Errorf("failed to get existing interactions: %s", err)
	}

	var removable []*discordgo.ApplicationCommand
	for _, cmd := range existing {
		if _, ok := b.Commands[cmd.Name]; !ok {
			removable = append(removable, cmd)
		}
	}

	for _, cmd := range removable {
		log.Debug().Str("command", cmd.Name).Msg("deleting command")
		if err := b.Session.ApplicationCommandDelete(b.AppID, "", cmd.ID, RequestOptions(ctx)); err != nil {
			return fmt.Errorf("failed to remove deprecated command %s(%s): %s", cmd.Name, cmd.ID, err)
		}
	}

	var updateable []*discordgo.ApplicationCommand
	for _, cmd := range b.Commands {
		if command, ok := cmd.(ApplicationCommandInteractionHandler); ok {
			spec := command.ApplicationCommand()
			log.Debug().Str("command", spec.Name).Msg("updating command")
			updateable = append(updateable, command.ApplicationCommand())
		}
	}

	if _, err := b.Session.ApplicationCommandBulkOverwrite(b.AppID, "", updateable, RequestOptions(ctx)); err != nil {
		return fmt.Errorf("failed to bulk update commands: %s", err)
	}

	return nil
}

func (b *Bot) GetGuild(ctx context.Context, id string) (*discordgo.Guild, error) {
	if guild, err := b.Session.State.Guild(id); err == nil {
		return guild, nil
	}

	guild, err := b.Session.Guild(id, RequestOptions(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get guild %s: %s", id, err)
	}

	if err := b.Session.State.GuildAdd(guild); err != nil {
		return nil, fmt.Errorf("failed to save guild %s to state: %s", id, err)
	}

	return guild, nil
}

func (b *Bot) GetEmoji(name string) *discordgo.Emoji {
	for _, guild := range b.Session.State.Guilds {
		for _, emoji := range guild.Emojis {
			if emoji.ID == name || emoji.Name == name {
				return emoji
			}
		}
	}

	return nil
}

func (b *Bot) GetEmojiForMessage(name, fallback string) string {
	if emoji := b.GetEmoji(name); emoji != nil {
		return emoji.MessageFormat()
	}

	return fallback
}

func (b *Bot) GetEmojiForReaction(name, fallback string) string {
	if emoji := b.GetEmoji(name); emoji != nil {
		return emoji.APIName()
	}

	return fallback
}

func (b *Bot) GetUserVoiceChannel(userID string) (guildID, channelID string) {
	for _, guild := range b.Session.State.Guilds {
		for _, state := range guild.VoiceStates {
			if state.UserID == userID {
				return guild.ID, state.ChannelID
			}
		}
	}

	return "", ""
}

func (b *Bot) JoinUserVoiceChannel(userID string) (*discordgo.VoiceConnection, error) {
	guildID, channelID := b.GetUserVoiceChannel(userID)

	if guildID == "" || channelID == "" {
		return nil, fmt.Errorf("no eligible voice channel found")
	}

	vc, err := b.Session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return nil, fmt.Errorf("failed to join voice channel %s: %s", vc.ChannelID, err)
	}

	return vc, nil
}

func (b *Bot) LeaveUserVoiceChannel(userID string) error {
	guildID, channelID := b.GetUserVoiceChannel(userID)
	if guildID == "" || channelID == "" {
		return nil
	}

	for _, vc := range b.Session.VoiceConnections {
		if vc.GuildID != guildID || vc.ChannelID != channelID {
			continue
		}

		if err := vc.Disconnect(); err != nil {
			return fmt.Errorf("failed to leave voice channel %s: %s", vc.ChannelID, err)
		}

		return nil
	}

	return fmt.Errorf("failed to find voice channel for user %s", userID)
}

func (b *Bot) SendMessageReaction(ctx context.Context, message *discordgo.Message, emoji, emojiFallback string) error {
	if err := b.Session.MessageReactionAdd(message.ChannelID, message.ID, b.GetEmojiForReaction(emoji, emojiFallback), RequestOptions(ctx)); err != nil {
		return fmt.Errorf("failed to react to message %s: %s", message.ID, err)
	}

	return nil
}

func (b *Bot) SendMessageReply(ctx context.Context, message *discordgo.Message, content string) error {
	if _, err := b.Session.ChannelMessageSendReply(message.ChannelID, content, message.Reference(), RequestOptions(ctx)); err != nil {
		return fmt.Errorf("failed to respond to message %s: %s", message.ID, err)
	}

	return nil
}

func (b *Bot) SendInteractionReply(ctx context.Context, interaction *discordgo.Interaction, response *discordgo.InteractionResponse) error {
	if err := b.Session.InteractionRespond(interaction, response, RequestOptions(ctx)); err != nil {
		return fmt.Errorf("failed to respond to interaction %s: %s", interaction.ID, err)
	}

	return nil
}

func (b *Bot) SendInteractionDeferral(ctx context.Context, interaction *discordgo.Interaction) error {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}

	return b.SendInteractionReply(ctx, interaction, &response)
}

func (b *Bot) SendInteractionMessageReply(ctx context.Context, interaction *discordgo.Interaction, content string) error {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	}

	return b.SendInteractionReply(ctx, interaction, &response)
}

func (b *Bot) SendInteractionEdit(ctx context.Context, interaction *discordgo.Interaction, edit *discordgo.WebhookEdit) error {
	if _, err := b.Session.InteractionResponseEdit(interaction, edit, RequestOptions(ctx)); err != nil {
		return fmt.Errorf("failed to edit interaction %s: %s", interaction.ID, err)
	}

	return nil
}
