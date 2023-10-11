package cmds

import (
	"context"
	"fmt"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/yt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	// _ discord.MessageCommandable     = (*Play)(nil)
	_ discord.ApplicationCommandable = (*Play)(nil)
)

type Play struct{ YouTube *yt.Client }

func (cmd Play) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "play",
		Description: "play a song",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "specify a url",
			},
		},
	}
}

type playOptions struct {
	url string
}

func (cmd Play) resolveOptions(opts []*discordgo.ApplicationCommandInteractionDataOption) playOptions {
	result := playOptions{}
	for _, opt := range opts {
		if opt.Name == "url" && opt.Type == discordgo.ApplicationCommandOptionString {
			result.url = opt.StringValue()
		}
	}

	return result
}

func (cmd Play) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	opts := cmd.resolveOptions(data.Options)

	if opts.url == "" {
		if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "must provide a url"); err != nil {
			log.Warn().Err(fmt.Errorf("failed to respond to interaction: %s", err)).Send()
		}

		return
	}

	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "🤔"); err != nil {
		log.Warn().Err(fmt.Errorf("failed to respond to interaction: %s", err)).Send()
	}

	item, err := cmd.YouTube.GetVideoByURL(ctx, opts.url)
	if err != nil {
		log.Warn().Err(err).Send()
		return
	}

	uploader := fmt.Sprintf("[%s](%s)", item.Snippet.ChannelTitle, item.ChannelURL())
	duration := item.Duration()

	edit := discordgo.WebhookEdit{
		Content: new(string),
		Embeds: &[]*discordgo.MessageEmbed{{
			Type:  discordgo.EmbedTypeRich,
			Title: item.Snippet.Title,
			URL:   item.VideoURL(),
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Uploader", Value: uploader, Inline: true},
				{Name: "Duration", Value: duration, Inline: true},
			},
		}},
	}

	if _, err := bot.Session.InteractionResponseEdit(event.Interaction, &edit, discord.WithRequestOptions(ctx)...); err != nil {
		log.Warn().Err(fmt.Errorf("failed to edit interaction: %s", err)).Send()
		return
	}
}
