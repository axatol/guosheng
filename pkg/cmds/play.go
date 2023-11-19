package cmds

import (
	"context"

	"github.com/axatol/guosheng/pkg/app"
	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/pkg/util"
	"github.com/axatol/guosheng/pkg/yt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.ApplicationCommandInteractionHandler = (*Play)(nil)
)

type Play struct{ App *app.App }

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
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "video_id",
				Description: "specify a video id",
			},
		},
	}
}

func (cmd Play) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	opts := resolveOptions(data.Options)
	videoID, ok := opts["video_id"].(string)
	if !ok {
		if url, ok := opts["url"].(string); ok {
			videoID = yt.GetVideoIDFromURL(url)
		}
	}

	if videoID == "" {
		if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "must provide an input"); err != nil {
			log.Warn().Err(err).Send()
		}

		return
	}

	if err := bot.SendInteractionMessageReply(ctx, event.Interaction, "🤔"); err != nil {
		log.Warn().Err(err).Send()
	}

	item, err := cmd.App.YouTube.GetVideoByID(ctx, videoID)
	if err != nil {
		log.Warn().Err(err).Send()
		return
	}

	uploader := util.MDLink(item.ChannelTitle, item.ChannelURL())
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
		log.Warn().Err(err).Send()
		return
	}

	go func() {
		user := event.User
		if user == nil {
			user = event.Member.User
		}

		vc, err := bot.JoinUserVoiceChannel(user.ID)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		song, err := cmd.App.GetSong(ctx, item)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		cmd.App.PlaySong(ctx, vc, song)
	}()
}
