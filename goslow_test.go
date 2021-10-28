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
		t.Run(c.name, func(t *testing.T) {
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
		})
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

func TestCancelledFunctionsArentCalled(t *testing.T) {
	cases := []struct {
		name string

		period  time.Duration
		timeout time.Duration
		context func() context.Context
	}{
		{"when already cancelled", time.Second, time.Millisecond, func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())

			cancel()

			return ctx
		}},
		{"when timeout exceeded", time.Second, time.Millisecond * 10, func() context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)

			cancel()

			return ctx
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sut := goslow.New(1, c.period)

			a := make(chan struct{})

			go sut.Do(c.context(), func() {
				close(a)
			})

			select {
			case <-a:
				t.Fail()
			case <-time.After(c.timeout):
			}
		})
	}
}
