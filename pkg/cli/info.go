package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

type Info struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Thumbnail      string `json:"thumbnail"`
	Uploader       string `json:"uploader"`
	UploaderID     string `json:"uploader_id"`
	UploaderURL    string `json:"uploader_url"`
	ChannelID      string `json:"channel_id"`
	ChannelURL     string `json:"channel_url"`
	Duration       int    `json:"duration"`
	DurationString string `json:"duration_string"`
	FormatID       string `json:"format_id"`
	Filename       string `json:"filename"`
}

func (e *Executor) GetInfo(id string) (*Info, error) {
	var info Info

	job := func(ctx context.Context) error {
		cmd := exec.CommandContext(ctx, e.YTDLPExecutable,
			fmt.Sprintf("https://youtube.com/watch?v=%s", id),
			"--cache-dir", e.CacheDirectory,
			"--dump-json",
		)

		raw, err := cmd.Output()
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok && len(err.Stderr) > 0 {
				log.Warn().Bytes("stderr", err.Stderr).Msg("stderr was not empty")
			}

			return fmt.Errorf("failed to execute '%s %s': %s", e.YTDLPExecutable, strings.Join(cmd.Args, " "), err)
		}

		if err := json.Unmarshal(raw, &info); err != nil {
			return fmt.Errorf("failed to parse output for %s: %s", id, err)
		}

		return nil
	}

	if err := e.execute(fmt.Sprintf("infojson:%s", id), job); err != nil {
		return nil, fmt.Errorf("failed to get info json for %s: %s", id, err)
	}

	return &info, nil
}
