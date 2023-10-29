package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func (e *Executor) Download(id string) (buffer []byte, err error) {
	job := func(ctx context.Context) (err error) {
		cmd := exec.CommandContext(ctx, e.YTDLPExecutable,
			fmt.Sprintf("https://youtube.com/watch?v=%s", id),
			"--cache-dir", e.CacheDirectory,
			"--abort-on-error",
			"--no-mark-watched",
			"--no-playlist",
			"--no-overwrites",
			"--continue",
			"--no-simulate",
			"--no-check-certificates",
			"--extract-audio",
			"--audio-format", "opus",
			"--prefer-free-formats",
			"--quiet",
			"--no-simulate",
			"--output", "-",
		)

		if buffer, err = cmd.Output(); err != nil {
			if err, ok := err.(*exec.ExitError); ok && len(err.Stderr) > 0 {
				log.Warn().Bytes("stderr", err.Stderr).Msg("stderr was not empty")
			}

			return fmt.Errorf("failed to execute '%s %s': %s", e.YTDLPExecutable, strings.Join(cmd.Args, " "), err)
		}

		return nil
	}

	if err = e.execute(fmt.Sprintf("download:%s", id), job); err != nil {
		return nil, fmt.Errorf("failed to download %s: %s", id, err)
	}

	return buffer, nil
}
