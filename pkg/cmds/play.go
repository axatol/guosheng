package cmds

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/axatol/guosheng/pkg/cache"
	"github.com/axatol/guosheng/pkg/cli"
	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/pkg/util"
	"github.com/axatol/guosheng/pkg/yt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.ApplicationCommandInteractionHandler = (*Play)(nil)
)

type Play struct {
	YouTube     *yt.Client
	CLI         *cli.Executor
	ObjectStore cache.ObjectStore
}

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

	item, err := cmd.YouTube.GetVideoByID(ctx, videoID)
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

		cacheKey := fmt.Sprintf("cache/%s", item.ID)
		if _, err := cmd.ObjectStore.Stat(ctx, cacheKey); err != nil {
			if err != cache.ErrObjectNotFound {
				log.Error().Err(err).Send()
				return
			}

			raw, err := cmd.CLI.Download(item.ID)
			if err != nil {
				log.Error().Err(err).Send()
				return
			}

			if _, err := cmd.ObjectStore.Put(ctx, cacheKey, raw, item.ToMap()); err != nil {
				log.Error().Err(err).Send()
				return
			}
		}

		raw, err := cmd.ObjectStore.Get(ctx, cacheKey)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		raw, err = cmd.CLI.Encode(item.ID, raw)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}

		if err := vc.Speaking(true); err != nil {
			log.Error().Err(fmt.Errorf("failed to start speaking: %s", err)).Send()
			return
		}

		defer func() {
			if err := vc.Speaking(false); err != nil {
				log.Error().Err(fmt.Errorf("failed to stop speaking: %s", err)).Send()
			}
		}()

		buffer := bytes.NewBuffer(raw)

		for {
			var frameLength int16
			if err := binary.Read(buffer, binary.LittleEndian, &frameLength); err != nil {
				log.Error().Err(fmt.Errorf("failed to read frame length: %s", err)).Send()
				return
			}

			frame := make([]byte, frameLength)
			if err := binary.Read(buffer, binary.LittleEndian, &frame); err != nil {
				if err != io.EOF && err != io.ErrUnexpectedEOF {
					log.Error().Err(fmt.Errorf("failed to read frame: %s", err)).Send()
				}

				return
			}

			vc.OpusSend <- frame
		}

		// TODO enqueue
	}()
}
