package nursery_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JamesOwenHall/nursery"
	"github.com/stretchr/testify/require"
)

var ErrCancelled = errors.New("cancelled")

func TestNurseryCtx(t *testing.T) {
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		n.Go(func() error {
			return ErrCancelled
		})
		n.Go(func() error {
			<-n.Ctx().Done()
			return nil
		})
	})

	require.Equal(t, ErrCancelled, err)
}

func TestNurserySendSuccess(t *testing.T) {
	var received int
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)
		n.Go(func() error {
			n.Send(ch, 5)
			return nil
		})
		n.Go(func() error {
			received = <-ch
			return nil
		})
	})

	require.NoError(t, err)
	require.Equal(t, 5, received)
}

func TestNurserySendCancelled(t *testing.T) {
	var received int
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)
		n.Go(func() error {
			n.Send(ch, 5)
			return nil
		})
		n.Go(func() error {
			return ErrCancelled
		})
	})

	require.Equal(t, ErrCancelled, err)
	require.Zero(t, received)
}

func TestNurseryRecvSuccess(t *testing.T) {
	var received int
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)
		n.Go(func() error {
			ch <- 5
			return nil
		})
		n.Go(func() error {
			n.Recv(ch, &received)
			return nil
		})
	})

	require.NoError(t, err)
	require.Equal(t, 5, received)
}

func TestNurseryRecvCancelled(t *testing.T) {
	var received int
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)
		n.Go(func() error {
			return ErrCancelled
		})
		n.Go(func() error {
			n.Recv(ch, &received)
			return nil
		})
	})

	require.Equal(t, ErrCancelled, err)
	require.Zero(t, received)
}

func BenchmarkNurseryManualSendAndRecv(b *testing.B) {
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)

		n.Go(func() error {
			defer close(ch)
			for i := 0; i < b.N; i++ {
				select {
				case ch <- i + 1:
				case <-n.Ctx().Done():
					return n.Ctx().Err()
				}
			}
			return nil
		})

		n.Go(func() error {
			for range ch {

			}
			return nil
		})
	})
	require.NoError(b, err)
}

func BenchmarkNurserySend(b *testing.B) {
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)

		n.Go(func() error {
			defer close(ch)
			for i := 0; i < b.N; i++ {
				n.Send(ch, i+1)
			}
			return nil
		})

		n.Go(func() error {
			for range ch {

			}
			return nil
		})
	})
	require.NoError(b, err)
}

func BenchmarkNurseryRecv(b *testing.B) {
	err := nursery.Supervise(context.Background(), func(n nursery.N) {
		ch := make(chan int)

		n.Go(func() error {
			defer close(ch)
			for i := 0; i < b.N; i++ {
				select {
				case ch <- i + 1:
				case <-n.Ctx().Done():
					return n.Ctx().Err()
				}
			}
			return nil
		})

		n.Go(func() error {
			for {
				var received int
				n.Recv(ch, &received)
				if received == 0 {
					return nil
				}
			}
		})
	})
	require.NoError(b, err)
}

func ExampleSupervise() {
	ctx := context.Background()
	err := nursery.Supervise(ctx, func(n nursery.N) {
		for i := 0; i < 5; i++ {
			i := i
			n.Go(func() error {
				return doWork(i)
			})
		}
	})

	if err != nil {
		// Handle error.
	}
}

var urls = []string{}

func ExampleSupervise_fanOut() {
	ctx := context.Background()
	err := nursery.Supervise(ctx, func(n nursery.N) {
		ch := make(chan string)

		// One goroutine to send work down the channel.
		n.Go(func() error {
			defer close(ch)
			for _, url := range urls {
				n.Send(ch, url)
			}
			return nil
		})

		// Many goroutines to received assigned work.
		for i := 0; i < 5; i++ {
			n.Go(func() error {
				for url := range ch {
					if err := crawl(url); err != nil {
						return err
					}
				}
				return nil
			})
		}
	})

	if err != nil {
		// Handle error.
	}
}

func doWork(_ int) error {
	return nil
}

func crawl(_ string) error {
	return nil
}
