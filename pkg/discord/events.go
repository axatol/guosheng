package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

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

func onConnect(session *discordgo.Session, event *discordgo.Connect) {
	log.Debug().Str("event", "Connect").Any("payload", event).Send()
}

func onDisconnect(session *discordgo.Session, event *discordgo.Disconnect) {
	log.Debug().Str("event", "Disconnect").Any("payload", event).Send()
}

// func onEvent(session *discordgo.Session, event *discordgo.Event) {
// 	log.Debug().Str("event", "Event").Any("payload", event).Send()
// }

func onInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	log.Debug().Str("event", "InteractionCreate").Any("payload", event).Send()
}

// func onMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
// 	log.Debug().Str("event", "MessageCreate").Any("payload", event).Send()
// }

// func onMessageReactionAdd(session *discordgo.Session, event *discordgo.MessageReactionAdd) {
// 	log.Debug().Str("event", "MessageReactionAdd").Any("payload", event).Send()
// }

func onRateLimit(session *discordgo.Session, event *discordgo.RateLimit) {
	log.Debug().Str("event", "RateLimit").Any("payload", event).Send()
}

func onReady(session *discordgo.Session, event *discordgo.Ready) {
	log.Debug().Str("event", "Ready").Any("payload", event).Send()

	usd := discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{{Details: "Plying the dildont"}},
	}

	if err := session.UpdateStatusComplex(usd); err != nil {
		log.Error().Err(fmt.Errorf("failed to update bot status: %s", err)).Send()
	}
}

// func onVoiceServerUpdate(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
// 	log.Debug().Str("event", "VoiceServerUpdate").Any("payload", event).Send()
// }

// func onVoiceStateUpdate(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
// 	log.Debug().Str("event", "VoiceStateUpdate").Any("payload", event).Send()
// }
