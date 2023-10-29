package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/axatol/go-utils/executil"
)

// from https://github.com/bwmarrin/dgvoice/blob/master/dgvoice.go
var (
	FFMPEGFormatPCM = "s16le"                 // pcm signed 16-bit little-endian
	OpusChannels    = 2                       // 1 for mono, 2 for stereo
	OpusFrameRate   = 48000                   // audio sampling rate
	OpusFrameSize   = 960                     // uint16 size of each audio frame
	OpusMaxBytes    = (OpusFrameSize * 2) * 2 // max size of opus data
)

func (e *Executor) Encode(id string, in []byte) (out []byte, err error) {
	job := func(ctx context.Context) (err error) {
		ctx, cancel := context.WithCancelCause(ctx)

		ffmpeg := exec.CommandContext(ctx, e.FFMPEGExecutable,
			"-i", "pipe:0",
			"-f", FFMPEGFormatPCM,
			"-ar", fmt.Sprint(OpusFrameRate),
			"-ac", fmt.Sprint(OpusChannels),
			"pipe:1",
		)

		ffmpeg.Stdin = bytes.NewReader(in)

		dca := exec.CommandContext(ctx, e.DCAExecutable,
			"-aa", "audio",
			"-ac", fmt.Sprint(OpusChannels),
			"-ar", fmt.Sprint(OpusFrameRate),
			"-as", fmt.Sprint(OpusFrameSize),
		)

		pipeline, err := executil.NewPipeline(ffmpeg, dca)
		if err != nil {
			return fmt.Errorf("failed to create pipeline: %s", err)
		}

		if out, err = pipeline.Execute(); err != nil {
			cancel(err)
			return fmt.Errorf("failed to execute pipeline: %s", err)
		}

		return nil
	}

	if err = e.execute(fmt.Sprintf("encode:%s", id), job); err != nil {
		return nil, fmt.Errorf("failed to encode %s: %s", id, err)
	}

	return out, nil
}
