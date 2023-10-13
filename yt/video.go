package yt

import (
	"context"
	"fmt"
	"net/url"

	"github.com/axatol/guosheng/pkg/util"
	"google.golang.org/api/youtube/v3"
)

type Video struct{ *youtube.Video }

func (v *Video) VideoURL() string {
	return fmt.Sprintf("https://youtube.com/watch?v=%s", v.Id)
}

func (v *Video) ChannelURL() string {
	return fmt.Sprintf("https://youtube.com/channel/%s", v.Snippet.ChannelId)
}

func (v *Video) Duration() *util.ISODuration {
	duration, err := util.ParseISODuration(v.ContentDetails.Duration)
	if err != nil {
		return nil
	}

	return duration
}

func (c *Client) GetVideo(ctx context.Context, id string) (*Video, error) {
	query := c.service.Videos.List([]string{"snippet", "contentDetails"}).Context(ctx).Id(id).MaxResults(1)
	response, err := query.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query video by id %s: %s", id, err)
	}

	if len(response.Items) < 1 {
		return nil, fmt.Errorf("video by id %s not found", id)
	}

	return &Video{response.Items[0]}, nil
}

func (c *Client) GetVideoByURL(ctx context.Context, rawURL string) (*Video, error) {
	u, err := url.Parse(NormaliseURL(rawURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %s: %s", rawURL, err)
	}

	if !u.Query().Has("v") {
		return nil, fmt.Errorf("video id not found in url %s", u.String())
	}

	return c.GetVideo(ctx, u.Query().Get("v"))
}
