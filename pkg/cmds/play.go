package cmds

import (
	"context"
	"fmt"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/pkg/yt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.ApplicationCommandInteractionHandler = (*Play)(nil)
)

type Play struct{ YouTube *yt.Client }

func (cmd Play) Name() string {
	return "play"
}

func (cmd Play) Description() string {
	return "Play a song"
}

func (cmd Play) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "specify a url",
			},
		},
	}
}

func (cmd Play) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	opts := resolveOptions(data.Options)
	url := opts["url"].(string)

	if url == "" {
		if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "must provide a url"); err != nil {
			log.Warn().Err(fmt.Errorf("failed to respond to interaction: %s", err)).Send()
		}

		return
	}

	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "🤔"); err != nil {
		log.Warn().Err(fmt.Errorf("failed to respond to interaction: %s", err)).Send()
	}

	item, err := cmd.YouTube.GetVideoByURL(ctx, url)
	if err != nil {
		log.Warn().Err(err).Send()
		return
	}

	uploader := fmt.Sprintf("[%s](%s)", item.ChannelTitle, item.ChannelURL())
	duration := item.Duration().String()
	if duration == "" {
		duration = "?"
	}

	edit := discordgo.WebhookEdit{
		Content: new(string),
		Embeds: &[]*discordgo.MessageEmbed{
			discord.NewMessageEmbed().
				SetTitle(item.Title).
				SetURL(item.VideoURL()).
				AddField("Uploader", uploader).
				AddField("Duration", duration).
				Embed(),
		},
	}

	if err := bot.SendInteractionEdit(ctx, event.Interaction, &edit); err != nil {
		log.Warn().Err(fmt.Errorf("failed to edit interaction: %s", err)).Send()
		return
	}
}
