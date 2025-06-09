package workerpool

import (
	"context"
	"sync"
)

type WorkerPool[T any] struct {
	idCounter      int
	idCounterMutex sync.Mutex
	workerMap      map[int]*Worker[T]
	mapMutex       sync.RWMutex
	task           Task[T]
	data           <-chan T
	errChan        chan<- error
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewWorkerPool[T any](task Task[T], parentCtx context.Context, data <-chan T, errChan chan<- error) *WorkerPool[T] {
	ctx, cancel := context.WithCancel(parentCtx)
	return &WorkerPool[T]{
		idCounter:      0,
		idCounterMutex: sync.Mutex{},
		workerMap:      make(map[int]*Worker[T]),
		mapMutex:       sync.RWMutex{},
		task:           task,
		data:           data,
		errChan:        errChan,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (wp *WorkerPool[T]) AddWorker() (int, error) {
	wp.idCounterMutex.Lock()
	defer wp.idCounterMutex.Unlock()
	wp.idCounter++
	currWorker := NewWorker(wp.idCounter, wp.task, wp.ctx)
	wp.mapMutex.Lock()
	defer wp.mapMutex.Unlock()
	wp.workerMap[wp.idCounter] = currWorker
	return wp.idCounter, currWorker.Run(wp.data, wp.errChan)
}

func (wp *WorkerPool[T]) RemoveWorker(id int) error {
	wp.mapMutex.Lock()
	defer wp.mapMutex.Unlock()
	currWorker, ok := wp.workerMap[id]
	if !ok {
		return NewDoesntExistsWorkerError(id)
	}
	currWorker.Stop()
	delete(wp.workerMap, id)
	return nil
}

func (wp *WorkerPool[T]) Stop() {
	wp.cancel()
}

func (wp *WorkerPool[T]) Wait() {
	wp.mapMutex.Lock()
	defer wp.mapMutex.Unlock()
	for key, _ := range wp.workerMap {
		w := wp.workerMap[key]
		w.Wait()
	}
}
