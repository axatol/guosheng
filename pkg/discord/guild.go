package discord

import (
	"github.com/bwmarrin/discordgo"
)

type Guild struct{ *discordgo.Guild }

func (g *Guild) GetEmoji(id string) *discordgo.Emoji {
	for _, e := range g.Emojis {
		if id == e.ID || id == e.Name {
			return e
		}
	}

	return nil
}
