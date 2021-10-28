package goslow_test

import (
	"context"
	"testing"
	"time"

	"github.com/wizhi/goslow"
)

func TestExhaustive(t *testing.T) {
	cases := []struct {
		name string

		max    int
		period time.Duration
		total  int
	}{
		{"WhenDivisible", 2, time.Millisecond, 6},
		{"WhenNotDivisible", 4, time.Millisecond, 6},
	}

	for _, c := range cases {
		sut := goslow.New(c.max, c.period)

		ctx := context.Background()

		seen := make(chan struct{}, c.max)

		for i := 0; i < c.total; i++ {
			go sut.Do(ctx, func() {
				seen <- struct{}{}
			})
		}

		for i := 0; i < c.total; i++ {
			<-seen
		}
	}
}

func TestFirstPeriodCallsFunctionsImmediately(t *testing.T) {
	period := time.Second

	sut := goslow.New(1, period)

	c := make(chan struct{})

	go sut.Do(context.Background(), func() {
		close(c)
	})

	select {
	case <-time.After(period):
		t.Fail()
	case <-c:
	}
}
