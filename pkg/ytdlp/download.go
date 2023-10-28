package ytdlp

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/axatol/go-utils/executil"
)

func (e *Executor) Download(id string) (buffer []byte, err error) {
	job := func(ctx context.Context) (err error) {
		ctx, cancel := context.WithCancelCause(ctx)

		ytdlpCmd := exec.CommandContext(ctx, e.YTDLPExecutable,
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

		ffmpegCmd := exec.CommandContext(ctx, e.FFMPEGExecutable,
			"-i", "pipe:0",
			"-f", "s16le",
			"-ar", "48000",
			"-ac", "2",
			"pipe:1",
		)

		dcaCmd := exec.CommandContext(ctx, e.DCAExecutable,
			"-aa", "audio",
			"-ac", "2",
			"-ar", "48000",
			"-as", "960",
		)

		buffer, err = executil.Pipeline(ytdlpCmd, ffmpegCmd, dcaCmd)
		if err != nil {
			cancel(fmt.Errorf("pipeline cancelled due to error: %s", err))
			return fmt.Errorf("failed to execute pipeline: %s", err)
		}

		return nil
	}

	if err = e.execute(fmt.Sprintf("download:%s", id), job); err != nil {
		return nil, fmt.Errorf("failed to download %s: %s", id, err)
	}

	return buffer, nil
}
