package util

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

type QueueItem struct {
	ID   string
	Work func(context.Context)
}

type Queue struct {
	concurrency int
	once        sync.Once
	items       chan QueueItem
}

func NewQueue(concurrency int) *Queue {
	return &Queue{
		concurrency: concurrency,
		once:        sync.Once{},
		items:       make(chan QueueItem),
	}
}

func (q *Queue) Start(ctx context.Context) {
	q.once.Do(func() {
		wg := sync.WaitGroup{}

		for i := 0; i < q.concurrency; i++ {
			wg.Add(1)

			go func(worker int) {
				for {
					select {
					case <-ctx.Done():
						// cancelled
						wg.Done()
						return

					case item := <-q.items:
						// got a job
						log.Debug().Int("worker", worker).Str("id", item.ID).Msg("executing job")
						item.Work(ctx)
					}
				}
			}(i)
		}

		wg.Wait()
	})
}

func (q *Queue) Enqueue(job QueueItem) {
	q.items <- job
}
