package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
)

type Worker[T any] struct {
	id      int
	task    Task[T]
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
}

func NewWorker[T any](id int, task Task[T], parentCtx context.Context) *Worker[T] {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Worker[T]{id, task, ctx, cancel, sync.WaitGroup{}, atomic.Bool{}}
}

func (w *Worker[T]) Run(data <-chan T, errChan chan<- error) error {
	if w.cancel == nil {
		return NewWorkerRestartError(w.id)
	}
	if !w.running.CompareAndSwap(false, true) {
		return NewAlreadyStartedError(w.id)
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.ctx.Done():
				return
			case curr, ok := <-data:
				if !ok {
					return
				}
				if err := w.task.Do(curr); err != nil {
					errChan <- err
				}
			}
		}
	}()
	return nil
}

func (w *Worker[T]) Wait() {
	w.wg.Wait()
	w.cancel = nil
}

func (w *Worker[T]) Stop() {
	if w.cancel != nil {
		w.cancel()
		w.cancel = nil
	}
}
