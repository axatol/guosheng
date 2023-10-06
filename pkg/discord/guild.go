package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type guildCtxKeyType string

var guildCtxKey guildCtxKeyType = "guild"

type Guild struct{ *discordgo.Guild }

func (g *Guild) GetEmoji(id string) *discordgo.Emoji {
	for _, e := range g.Emojis {
		if id == e.ID || id == e.Name {
			return e
		}
	}

	return nil
}

func GuildFromContext(ctx context.Context) *Guild {
	if value, ok := ctx.Value(guildCtxKey).(*Guild); ok {
		return value
	}

	return nil
}

func (g *Guild) InContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, guildCtxKey, g)
}
