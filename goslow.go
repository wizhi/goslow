package goslow

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"
	"time"
)

func New(max int, period time.Duration) *slow {
	return &slow{
		max:    int64(max),
		period: period,
	}
}

type slow struct {
	max    int64
	period time.Duration

	current int64
	waiting list.List
	ticker  *time.Ticker

	once sync.Once
}

func (t *slow) Do(ctx context.Context, f func()) error {
	t.once.Do(func() {
		t.ticker = time.NewTicker(t.period)
	})

	if atomic.AddInt64(&t.current, 1) <= t.max {
		f()

		return nil
	}

	ready := make(chan struct{})

	t.waiting.PushBack(ready)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.ticker.C:
			atomic.StoreInt64(&t.current, 0)

			next := t.waiting.Front()

			for i := int64(0); i < t.max && next != nil; i++ {
				t.waiting.Remove(next)

				close(next.Value.(chan struct{}))

				next = t.waiting.Front()
			}

			continue
		case <-ready:
			atomic.AddInt64(&t.current, 1)

			f()

			return nil
		}
	}
}
