package cli

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/axatol/guosheng/pkg/util"
)

type Executor struct {
	YTDLPExecutable  string
	FFMPEGExecutable string
	DCAExecutable    string
	Concurrency      int
	CacheDirectory   string
	queue            *util.Queue
}

func (e *Executor) Listen(ctx context.Context) error {
	if e.YTDLPExecutable == "" {
		e.YTDLPExecutable = "yt-dlp"
	}

	if e.FFMPEGExecutable == "" {
		e.FFMPEGExecutable = "ffmpeg"
	}

	if e.DCAExecutable == "" {
		e.DCAExecutable = "dca"
	}

	if e.queue == nil {
		e.queue = util.NewQueue(e.Concurrency)
	}

	if e.Concurrency < 1 {
		return fmt.Errorf("concurrency must be greater than 0, got: %d", e.Concurrency)
	}

	if _, err := os.Stat(e.CacheDirectory); err != nil {
		return fmt.Errorf("could not stat cache directory %s: %s", e.CacheDirectory, err)
	}

	go e.queue.Start(ctx)
	return nil
}

func (e *Executor) execute(id string, job func(context.Context) error) (err error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	worker := func(ctx context.Context) {
		defer wg.Done()
		err = job(ctx)
	}

	e.queue.Enqueue(util.QueueItem{
		ID:   id,
		Work: worker,
	})

	return err
}
