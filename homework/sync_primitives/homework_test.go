package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type RWMutex struct {
	mtx  sync.Mutex
	cnd  sync.Cond
	rcnt atomic.Int32
	wcnt atomic.Int32
}

func NewRWMutex() RWMutex {
	return RWMutex{
		cnd: sync.Cond{L: &sync.Mutex{}},
	}
}

func (m *RWMutex) Lock() {
	m.wcnt.Add(1)
	m.mtx.Lock()
}

func (m *RWMutex) Unlock() {
	m.mtx.Unlock()
	m.wcnt.Add(-1)
	m.cnd.Broadcast()
}

func (m *RWMutex) RLock() {
	m.cnd.L.Lock()
	defer m.cnd.L.Unlock()

	for m.wcnt.Load() != 0 {
		m.cnd.Wait()
	}

	if m.rcnt.CompareAndSwap(0, 1) {
		m.mtx.Lock()
	} else {
		m.rcnt.Add(1)
	}
}

func (m *RWMutex) RUnlock() {
	if m.rcnt.CompareAndSwap(1, 0) {
		m.mtx.Unlock()
	} else {
		m.rcnt.Add(-1)
	}
}

func TestRWMutexWithWriter(t *testing.T) {
	var mutex RWMutex = NewRWMutex()
	mutex.Lock() // writer

	var mutualExclusionWithWriter atomic.Bool
	mutualExclusionWithWriter.Store(true)
	var mutualExclusionWithReader atomic.Bool
	mutualExclusionWithReader.Store(true)

	go func() {
		mutex.Lock() // another writer
		mutualExclusionWithWriter.Store(false)
	}()

	go func() {
		mutex.RLock() // another reader
		mutualExclusionWithReader.Store(false)
	}()

	time.Sleep(time.Second)
	assert.True(t, mutualExclusionWithWriter.Load())
	assert.True(t, mutualExclusionWithReader.Load())
}

func TestRWMutexWithReaders(t *testing.T) {
	var mutex RWMutex = NewRWMutex()
	mutex.RLock() // reader

	var mutualExclusionWithWriter atomic.Bool
	mutualExclusionWithWriter.Store(true)

	go func() {
		mutex.Lock() // another writer
		mutualExclusionWithWriter.Store(false)
	}()

	time.Sleep(time.Second)
	assert.True(t, mutualExclusionWithWriter.Load())
}

func TestRWMutexMultipleReaders(t *testing.T) {
	var mutex RWMutex = NewRWMutex()
	mutex.RLock() // reader

	var readersCount atomic.Int32
	readersCount.Add(1)

	go func() {
		mutex.RLock() // another reader
		readersCount.Add(1)
	}()

	go func() {
		mutex.RLock() // another reader
		readersCount.Add(1)
	}()

	time.Sleep(time.Second)
	assert.Equal(t, int32(3), readersCount.Load())
}

func TestRWMutexWithWriterPriority(t *testing.T) {
	var mutex RWMutex = NewRWMutex()
	mutex.RLock() // reader

	var mutualExclusionWithWriter atomic.Bool
	mutualExclusionWithWriter.Store(true)
	var readersCount atomic.Int32
	readersCount.Add(1)

	go func() {
		mutex.Lock() // another writer is waiting for reader
		mutualExclusionWithWriter.Store(false)
	}()

	time.Sleep(time.Second)

	go func() {
		mutex.RLock() // another reader is waiting for a higher priority writer
		readersCount.Add(1)
	}()

	go func() {
		mutex.RLock() // another reader is waiting for a higher priority writer
		readersCount.Add(1)
	}()

	time.Sleep(time.Second)

	assert.True(t, mutualExclusionWithWriter.Load())
	assert.Equal(t, int32(1), readersCount.Load())
}

func TestRWMutexTwoWriters(t *testing.T) {
	var mutex RWMutex = NewRWMutex()

	var writerCounter atomic.Int32

	go func() {
		mutex.Lock()
		defer mutex.Unlock()

		time.Sleep(time.Second) // some work
		writerCounter.Add(1)
	}()

	go func() {
		mutex.Lock()
		defer mutex.Unlock()

		time.Sleep(time.Second) // some work
		writerCounter.Add(1)
	}()

	time.Sleep(3 * time.Second)

	assert.Equal(t, int32(2), writerCounter.Load())
}

func TestRWMutexTwoWritersDeadlock(t *testing.T) {
	var mutex RWMutex = NewRWMutex()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		mutex.Lock()
		defer mutex.Unlock()

		time.Sleep(200 * time.Millisecond) // some work
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		mutex.Lock()
		defer mutex.Unlock()

		time.Sleep(200 * time.Millisecond) // some work
	}()

	wg.Wait()
}
