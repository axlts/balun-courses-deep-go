package main

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type WorkerPool struct {
	in chan func()
	wg sync.WaitGroup

	stopped atomic.Bool
}

func NewWorkerPool(workersNumber, taskBuffer int) *WorkerPool {
	wp := &WorkerPool{
		in: make(chan func(), taskBuffer),
	}

	wp.wg.Add(workersNumber)
	for i := 0; i < workersNumber; i++ {
		go func() {
			defer wp.wg.Done()
			for task := range wp.in {
				task()
			}
		}()
	}

	return wp
}

// Return an error if the pool is full.
func (wp *WorkerPool) AddTask(task func()) error {
	if wp.stopped.Load() {
		return errors.New("worker pool is stopped")
	}
	if task == nil {
		return errors.New("task is nil")
	}

	select {
	case wp.in <- task:
		return nil
	default:
		return errors.New("task pool is full")
	}
}

// Shutdown all workers and wait for all
// tasks in the pool to complete.
func (wp *WorkerPool) Shutdown() {
	if !wp.stopped.CompareAndSwap(false, true) {
		return
	}
	close(wp.in)
	wp.wg.Wait()
}

func TestWorkerPool(t *testing.T) {
	var counter atomic.Int32
	task := func() {
		time.Sleep(time.Millisecond * 500)
		counter.Add(1)
	}

	pool := NewWorkerPool(2, 3)
	_ = pool.AddTask(task)
	_ = pool.AddTask(task)
	_ = pool.AddTask(task)

	time.Sleep(time.Millisecond * 600)
	assert.Equal(t, int32(2), counter.Load())

	time.Sleep(time.Millisecond * 600)
	assert.Equal(t, int32(3), counter.Load())

	_ = pool.AddTask(task)
	_ = pool.AddTask(task)
	_ = pool.AddTask(task)
	pool.Shutdown() // wait tasks

	assert.Equal(t, int32(6), counter.Load())
}
