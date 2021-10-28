package goslow

import (
	"context"
	"sync"
	"time"
)

func New(max int, period time.Duration) *slow {
	queue := make(chan chan struct{}, max)
	ticker := time.NewTicker(period)

	ticker.Stop()

	current := 0

	go func() {
		for f := range queue {
			current++

			if current > max {
				<-ticker.C

				current -= max
			}

			close(f)
		}
	}()

	return &slow{
		period: period,
		queue:  queue,
		ticker: ticker,
	}
}

type slow struct {
	period time.Duration

	queue  chan chan struct{}
	ticker *time.Ticker

	once sync.Once
}

func (s *slow) Do(ctx context.Context, f func()) error {
	ready := make(chan struct{})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.queue <- ready:
	}

	s.once.Do(func() {
		s.ticker.Reset(s.period)
	})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ready:
		f()

		return nil
	}
}
