package yt

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	service *youtube.Service
}

func New(ctx context.Context, key string) (*Client, error) {
	service, err := youtube.NewService(ctx, option.WithAPIKey(key))
	if err != nil {
		return nil, fmt.Errorf("failed to create youtube service: %s", err)
	}

	return &Client{service}, nil
}

func NormaliseURL(raw string) string {
	raw = strings.Replace(raw, "music.youtube.com", "youtube.com", 1)
	raw = strings.Replace(raw, "youtu.be", "youtube.com", 1)
	raw = strings.Replace(raw, "/v/", "/watch?v=", 1)
	raw = strings.Replace(raw, "/watch#", "/watch?", 1)
	raw = strings.Replace(raw, "/youtube.com/shorts/", "youtube.com/watch?v=", 1)
	return raw
}

func GetVideoIDFromURL(input string) string {
	url, err := url.Parse(input)
	if err != nil {
		return ""
	}

	return url.Query().Get("q")
}
