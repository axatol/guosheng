package cmds

import (
	"context"
	"fmt"
	"strings"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/yt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

var (
	_ discord.ApplicationCommandInteractionHandler = (*Search)(nil)
	_ discord.MessageComponentInteractionHandler   = (*Search)(nil)
)

type Search struct{ YouTube *yt.Client }

func (cmd Search) Name() string {
	return "search"
}

func (cmd Search) Description() string {
	return "Search for a youtube song to play"
}

func (cmd Search) ApplicationCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        cmd.Name(),
		Description: cmd.Description(),
		Options: []*discordgo.ApplicationCommandOption{{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "query",
			Description: "Search terms",
			Required:    true,
		}},
	}
}

func (cmd Search) OnApplicationCommand(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.ApplicationCommandInteractionData) {
	opts := resolveOptions(data.Options)
	query := opts["query"].(string)
	searchResults, err := cmd.YouTube.SearchVideo(ctx, query, 5)
	log.Trace().Str("query", query).Any("search_results", searchResults).Send()
	if err != nil {
		response := discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Title:   "Failed to search",
				Content: err.Error(),
			},
		}

		if err := bot.SendInteractionReply(ctx, event.Interaction, &response); err != nil {
			log.Warn().Err(err).Send()
		}

		return
	}

	embed := discord.NewMessageEmbed().
		SetTitle("Search results").
		SetType(discordgo.EmbedTypeRich)
	for i, item := range searchResults {
		embed.AddField(
			fmt.Sprintf("%d. %s", i+1, item.Title),
			fmt.Sprintf("uploaded by [%s](%s), view on [youtube](%s)", item.ChannelTitle, item.ChannelURL(), item.VideoURL()),
			false,
		)
	}

	options := make([]discordgo.SelectMenuOption, len(searchResults))
	for i, item := range searchResults {
		options[i] = discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%d. %s", i+1, item.Title),
			Description: fmt.Sprintf("uploaded by %s", item.ChannelTitle),
			Value:       item.ID,
		}
	}

	customID := fmt.Sprintf("%s:%s", cmd.Name(), event.ID)
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed.Embed()},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							Placeholder: "Select items to enqueue",
							CustomID:    customID,
							MenuType:    discordgo.StringSelectMenu,
							Options:     options,
						},
					},
				},
			},
		},
	}

	if err := bot.SendInteractionReply(ctx, event.Interaction, &response); err != nil {
		log.Warn().Err(err).Send()
		return
	}
}

func (cmd Search) OnMessageComponent(ctx context.Context, bot *discord.Bot, event *discordgo.InteractionCreate, data *discordgo.MessageComponentInteractionData) {
	customID := strings.Split(data.CustomID, ":")[1]
	videoIDs := data.Values

	log := log.With().
		Str("custom_id", customID).
		Strs("video_ids", videoIDs).
		Logger()

	if err := bot.SendInteractionDeferral(ctx, event.Interaction); err != nil {
		log.Warn().Err(err).Send()
	}

	results, err := cmd.YouTube.GetVideosByIDs(ctx, videoIDs...)
	if err != nil {
		log.Warn().Err(err).Send()
	}

	embeds := make([]*discordgo.MessageEmbed, len(results))
	for i, result := range results {
		embeds[i] = result.Embed()
	}

	edit := discordgo.WebhookEdit{
		Embeds:     &embeds,
		Components: &[]discordgo.MessageComponent{},
	}

	if err := bot.SendInteractionEdit(ctx, event.Interaction, &edit); err != nil {
		log.Warn().Err(err).Send()
	}
}
