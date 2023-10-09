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
	Guilds   map[string]Guild
	Emojis   map[string]discordgo.Emoji
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
		Guilds:     make(map[string]Guild),
		Emojis:     make(map[string]discordgo.Emoji),
		Commands:   make(map[string]any),
	}

	bot.Session.Identify.Intents = discordIntents
	bot.Session.AddHandler(bot.onEvent)
	bot.Session.AddHandler(bot.onGuildCreate)
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

	return b.Session.Close()
}

func (b *Bot) RegisterInteractions(ctx context.Context) error {
	existing, err := b.Session.ApplicationCommands(b.AppID, "", WithRequestOptions(ctx)...)
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
		if err := b.Session.ApplicationCommandDelete(b.AppID, "", cmd.ID, WithRequestOptions(ctx)...); err != nil {
			return fmt.Errorf("failed to remove deprecated command %s(%s): %s", cmd.Name, cmd.ID, err)
		}
	}

	var updateable []*discordgo.ApplicationCommand
	for _, cmd := range b.Commands {
		if command, ok := cmd.(ApplicationCommandable); ok {
			updateable = append(updateable, command.ApplicationCommand())
		}
	}

	if _, err := b.Session.ApplicationCommandBulkOverwrite(b.AppID, "", updateable, WithRequestOptions(ctx)...); err != nil {
		return fmt.Errorf("failed to bulk update commands: %s", err)
	}

	return nil
}

func (b *Bot) GetGuild(ctx context.Context, id string) (*Guild, error) {
	if guild, ok := b.Guilds[id]; ok {
		return &guild, nil
	}

	discordGuild, err := b.Session.Guild(id, WithRequestOptions(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild %s: %s", id, err)
	}

	for _, e := range discordGuild.Emojis {
		if e != nil {
			b.Emojis[e.Name] = *e
		}
	}

	guild := Guild{discordGuild}
	b.Guilds[id] = guild
	return &guild, nil
}
