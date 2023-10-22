package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/axatol/go-utils/contextutil"
	"github.com/axatol/guosheng/pkg/cmds"
	"github.com/axatol/guosheng/pkg/config"
	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/pkg/server"
	"github.com/axatol/guosheng/pkg/server/handlers"
	"github.com/axatol/guosheng/pkg/yt"
	"github.com/rs/zerolog/log"
)

var exitCode = 0

func main() {
	defer os.Exit(exitCode)

	config.Configure()

	if config.PrintVersion {
		config.Version().Info().Send()
		return
	}

	ctx := context.Background()
	ctx = contextutil.WithInterrupt(ctx)
	ctx, cancel := context.WithCancelCause(ctx)

	bot := runBot(ctx, cancel)
	server := runServer(ctx, cancel)

	<-ctx.Done()
	if err := context.Cause(ctx); err != nil && err != context.Canceled {
		log.Error().Err(fmt.Errorf("context canceled: %s", err)).Send()
		exitCode = 1
	}

	cleanup(bot, server)
}

func runBot(ctx context.Context, cancel context.CancelCauseFunc) *discord.Bot {
	yt, err := yt.New(ctx, config.YouTubeAPIKey)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	botOpts := discord.BotOptions{
		AppID:         config.DiscordAppID,
		BotToken:      config.DiscordBotToken,
		MessagePrefix: config.DiscordMessagePrefix,
	}

	bot, err := discord.NewBot(botOpts)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	shutdown := func() {
		cancel(fmt.Errorf("received shutdown command"))
	}

	bot.RegisterCommand(ctx, cmds.Shutdown{Shutdown: shutdown})
	bot.RegisterCommand(ctx, cmds.Beep{})
	bot.RegisterCommand(ctx, cmds.Join{})
	bot.RegisterCommand(ctx, cmds.Leave{})
	bot.RegisterCommand(ctx, cmds.Play{YouTube: yt})
	bot.RegisterCommand(ctx, cmds.Search{YouTube: yt})
	// should be last
	bot.RegisterCommand(ctx, cmds.Help{Commands: bot.Commands})

	if err := bot.RegisterInteractions(ctx); err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := bot.Open(ctx, time.Second*10); err != nil {
		log.Fatal().Err(err).Send()
	}

	return bot
}

func runServer(ctx context.Context, cancel context.CancelCauseFunc) *http.Server {
	router := server.NewRouter(config.ServerAddress)
	router.Get("/api/ping", handlers.Ping)
	router.Get("/api/health", handlers.Health)

	server := http.Server{Addr: config.ServerAddress, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cancel(err)
		}
	}()

	return &server
}

func cleanup(bot *discord.Bot, server *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ctx, cancel := context.WithCancelCause(context.Background())
	queue := sync.WaitGroup{}

	queue.Add(1)
	go func() {
		defer queue.Done()
		if err := bot.Close(); err != nil {
			log.Warn().Err(fmt.Errorf("failed to shutdown bot: %s", err)).Send()
		}
	}()

	queue.Add(1)
	go func() {
		defer queue.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Warn().Err(fmt.Errorf("failed to shutdown server: %s", err)).Send()
		}
	}()

	// wait for tasks to complete successfully
	go func() {
		queue.Wait()
		cancel(nil)
	}()

	// wait for failure to clean up
	deadline := time.NewTimer(time.Second * 10)
	select {
	case <-sig:
		// forceful interrupt
		cancel(fmt.Errorf("received interrupt"))
	case <-deadline.C:
		// timeout
		cancel(fmt.Errorf("context deadline exceeded"))
	case <-ctx.Done():
		// parent context finished
	}

	if err := context.Cause(ctx); err != nil {
		log.Error().Err(fmt.Errorf("cleanup canceled: %s", err)).Send()
		exitCode = 1
	}
}
