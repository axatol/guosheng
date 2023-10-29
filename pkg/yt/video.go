package yt

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/axatol/guosheng/pkg/discord"
	"github.com/axatol/guosheng/pkg/util"
	"github.com/bwmarrin/discordgo"
)

type Video struct {
	ID           string `json:"id"`
	ETag         string `json:"etag"`
	Title        string `json:"title"`
	ChannelID    string `json:"channel_id"`
	ChannelTitle string `json:"channel_title"`
	DurationRaw  string `json:"duration"`
}

func (v *Video) ToMap() map[string]string {
	return map[string]string{
		"id":            v.ID,
		"etag":          v.ETag,
		"title":         v.Title,
		"channel_id":    v.ChannelID,
		"channel_title": v.ChannelTitle,
		"duration":      v.DurationRaw,
	}
}

func (v *Video) VideoURL() string {
	return fmt.Sprintf("https://youtube.com/watch?v=%s", v.ID)
}

func (v *Video) ChannelURL() string {
	return fmt.Sprintf("https://youtube.com/channel/%s", v.ChannelID)
}

func (v *Video) Duration() *util.ISODuration {
	if v.DurationRaw == "" {
		return nil
	}

	duration, err := util.ParseISODuration(v.DurationRaw)
	if err != nil {
		return nil
	}

	return duration
}

func (v *Video) AsMessageEmbed() *discordgo.MessageEmbed {
	uploader := util.MDLink(v.ChannelTitle, v.ChannelURL())
	duration := v.Duration().String()
	if duration == "" {
		duration = "?"
	}

	return discord.NewMessageEmbed().
		SetTitle(v.Title).
		SetURL(v.VideoURL()).
		AddField("Uploader", uploader).
		AddField("Duration", duration).
		Embed()
}

func (c *Client) GetVideosByIDs(ctx context.Context, ids ...string) ([]Video, error) {
	query := c.service.Videos.
		List([]string{"snippet", "contentDetails"}).
		Context(ctx).
		Id(ids...).
		MaxResults(int64(len(ids)))

	response, err := query.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query video by id %s: %s", strings.Join(ids, ","), err)
	}

	if len(response.Items) < 1 {
		return nil, fmt.Errorf("video by id %s not found", strings.Join(ids, ","))
	}

	result := make([]Video, len(response.Items))
	for i, item := range response.Items {
		result[i] = Video{
			ID:           item.Id,
			ETag:         item.Etag,
			Title:        item.Snippet.Title,
			ChannelID:    item.Snippet.ChannelId,
			ChannelTitle: item.Snippet.ChannelTitle,
			DurationRaw:  item.ContentDetails.Duration,
		}
	}

	return result, nil
}

func (c *Client) GetVideoByID(ctx context.Context, id string) (*Video, error) {
	results, err := c.GetVideosByIDs(ctx, id)
	if err != nil {
		return nil, err
	}

	return &results[0], nil
}

func (c *Client) GetVideoByURL(ctx context.Context, rawURL string) (*Video, error) {
	u, err := url.Parse(NormaliseURL(rawURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %s", rawURL, err)
	}

	if !u.Query().Has("v") {
		return nil, fmt.Errorf("video id not found in url %s", u.String())
	}

	return c.GetVideoByID(ctx, u.Query().Get("v"))
}

func (c *Client) SearchVideo(ctx context.Context, search string, limit int64) ([]Video, error) {
	query := c.service.Search.List([]string{"snippet"}).Context(ctx).Q(search).SafeSearch("none").Type("video").MaxResults(limit)
	response, err := query.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search video with query %s: %s", search, err)
	}

	if len(response.Items) < 1 {
		return nil, nil
	}

	result := make([]Video, len(response.Items))
	for i, item := range response.Items {
		result[i] = Video{
			ID:           item.Id.VideoId,
			ETag:         item.Etag,
			Title:        item.Snippet.Title,
			ChannelID:    item.Snippet.ChannelId,
			ChannelTitle: item.Snippet.ChannelTitle,
			DurationRaw:  "",
		}
	}

	return result, nil
}
