package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func WithRequestOptions(ctx context.Context, reason ...string) []discordgo.RequestOption {
	opts := []discordgo.RequestOption{
		discordgo.WithContext(ctx),
		discordgo.WithRetryOnRatelimit(true),
		// TODO discordgo.WithClient(defaultRestClient),
	}

	if len(reason) == 1 {
		opts = append(opts, discordgo.WithAuditLogReason(reason[0]))
	}

	return opts
}
