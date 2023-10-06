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

	"github.com/axatol/guosheng/pkg/config"
	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/pkg/server"
	"github.com/axatol/guosheng/pkg/server/handlers"
	"github.com/rs/zerolog/log"
)

var exitCode = 0

func main() {
	defer os.Exit(exitCode)
	config.Configure()
	config.Version().Info().Send()

	ctx := context.Background()

	bot, err := discord.NewBot(config.DiscordBotToken, config.DiscordBotPrefix)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	if err := bot.Open(ctx); err != nil {
		log.Fatal().Err(err).Send()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		<-sig
		cancel(fmt.Errorf("received interrupt, gracefully shutting down"))
	}()

	router := server.NewRouter(config.ServerAddress)
	router.Get("/api/ping", handlers.Ping)
	router.Get("/api/health", handlers.Health)
	router.Get("/api/shutdown", handlers.Shutdown(cancel))
	server := http.Server{Addr: config.ServerAddress, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cancel(err)
		} else {
			cancel(nil)
		}
	}()

	<-ctx.Done()
	if err := context.Cause(ctx); err != nil {
		log.Error().Err(fmt.Errorf("context canceled: %s", err)).Send()
		exitCode = 1
	}

	cleanup(bot, &server)
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
	deadline := time.NewTimer(time.Second * 7)
	select {
	case <-sig:
		// forceful interrupt
		cancel(fmt.Errorf("received interrupt, forcefully shutting down"))
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
