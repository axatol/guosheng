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

type Bot struct {
	session *discordgo.Session
	guilds  map[string]Guild
}

func NewBot(token, prefix string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("discord bot token is required")
	}

	session, err := discordgo.New(fmt.Sprintf("Bot %s", config.DiscordBotToken))
	if err != nil {
		log.Fatal().Err(fmt.Errorf("failed to start discord session: %s", err)).Send()
	}

	bot := Bot{
		session: session,
		guilds:  map[string]Guild{},
	}

	bot.session.ShouldReconnectOnError = true
	bot.session.ShouldRetryOnRateLimit = false
	bot.session.Identify.Intents = discordIntents

	bot.session.AddHandler(onConnect)
	bot.session.AddHandler(onDisconnect)
	bot.session.AddHandler(onEvent)
	bot.session.AddHandler(onGuildCreate(bot))
	bot.session.AddHandler(onInteractionCreate)
	bot.session.AddHandler(onMessageCreate(prefix))
	bot.session.AddHandler(onMessageReactionAdd)
	bot.session.AddHandler(onRateLimit)
	bot.session.AddHandler(onReady)
	bot.session.AddHandler(onVoiceServerUpdate)
	bot.session.AddHandler(onVoiceStateUpdate)

	return &bot, nil
}

func waitForReady(ctx context.Context, session *discordgo.Session) error {
	wait := make(chan struct{}, 1)
	session.AddHandlerOnce(func(s *discordgo.Session, e *discordgo.Ready) {
		wait <- struct{}{}
	})

	if err := session.Open(); err != nil {
		return fmt.Errorf("failed to open discord session: %s", err)
	}

	deadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	select {
	case <-deadline.Done():
		return fmt.Errorf("failed to connect to discord: %s", deadline.Err())
	case <-wait:
		return nil
	}
}

func (b *Bot) GetGuild(ctx context.Context, id string) (*Guild, error) {
	if guild, ok := b.guilds[id]; ok {
		return &guild, nil
	}

	discordGuild, err := b.session.Guild(id, discordgo.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get guild %s: %s", id, err)
	}

	guild := Guild{discordGuild}
	b.guilds[id] = guild
	return &guild, nil
}

func (b *Bot) Open(ctx context.Context) error {
	log.Debug().Msg("waiting for discord bot to be ready")
	if err := waitForReady(ctx, b.session); err != nil {
		return fmt.Errorf("failed to ready discord session: %s", err)
	}

	log.Debug().Msg("discord bot session established")
	return nil
}

func (b *Bot) Close() error {
	return b.session.Close()
}
