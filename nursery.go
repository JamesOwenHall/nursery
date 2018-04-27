// Package nursery provides a control flow mechanism for guaranteeing that
// multiple goroutines will exit before continuing execution.
//
// For an explanation of the concept, see
// https://vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/
package nursery

import (
	"context"
	"reflect"
	"runtime"
	"sync"
)

// N provides common functions for spawning, communicating with, and waiting
// for goroutines.
type N interface {
	// Go spawns a new goroutine. If the error returned by fn is not nil, then
	// the nursery's context will be cancelled, along with all pending Send and
	// Recv calls.
	Go(fn func() error)
	// Ctx returns the context associated with the nursery.
	Ctx() context.Context
	// Send sends the value v down the channel c. If the nursery's context is
	// cancelled, the goroutine will call all deferred functions and exit.
	Send(c interface{}, v interface{})
	// Recv receives a value from the channel c and stores it in v. If the
	// nursery's context is cancelled, the goroutine will call all deferred
	// functions and exit.
	Recv(c interface{}, v interface{})
}

// Supervise provides a nursery to spawn goroutines and wait for them all to
// exit. It will return the first non-nil error returned from any goroutine
// started with the nursery n.
func Supervise(ctx context.Context, fn func(n N)) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	n := &nursery{
		ctx:    ctx,
		cancel: cancel,
	}

	fn(n)
	n.wg.Wait()
	return n.err
}

type nursery struct {
	ctx    context.Context
	cancel func()
	once   sync.Once
	err    error
	wg     sync.WaitGroup
}

func (n *nursery) Go(fn func() error) {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := fn(); err != nil {
			n.once.Do(func() {
				n.cancel()
				n.err = err
			})
		}
	}()
}

func (n *nursery) Ctx() context.Context {
	return n.ctx
}

func (n *nursery) Send(c interface{}, v interface{}) {
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.ctx.Done()),
		},
		{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(c),
			Send: reflect.ValueOf(v),
		},
	}

	chosen, _, _ := reflect.Select(cases)
	if chosen == 0 {
		runtime.Goexit()
	}
}

func (n *nursery) Recv(c interface{}, v interface{}) {
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.ctx.Done()),
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(c),
		},
	}

	chosen, received, _ := reflect.Select(cases)
	if chosen == 0 {
		runtime.Goexit()
	}

	dest := reflect.ValueOf(v)
	dest.Elem().Set(received)
}
