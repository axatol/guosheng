package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func RequestOptions(ctx context.Context, reason ...string) discordgo.RequestOption {
	return func(cfg *discordgo.RequestConfig) {
		discordgo.WithContext(ctx)(cfg)
		discordgo.WithRetryOnRatelimit(true)(cfg)
		discordgo.WithRestRetries(3)

		if len(reason) == 1 {
			discordgo.WithAuditLogReason(reason[0])(cfg)
		}
	}
}
