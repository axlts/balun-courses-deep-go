package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type RWMutex struct {
	mtx        sync.Mutex
	rcnt, wcnt atomic.Int64
}

func (m *RWMutex) Lock() {
	m.wcnt.Add(1)
	m.mtx.Lock()
}

func (m *RWMutex) Unlock() {
	m.wcnt.Add(-1)
	m.mtx.Unlock()
}

func (m *RWMutex) RLock() {
	if m.rcnt.CompareAndSwap(0, 1) {
		m.mtx.Lock()
		return
	} else if m.wcnt.Load() > 0 {
		m.mtx.Lock()
	}
	m.rcnt.Add(1)
}

func (m *RWMutex) RUnlock() {
	if m.rcnt.CompareAndSwap(1, 0) {
		m.mtx.Unlock()
		return
	}
	m.rcnt.Add(-1)
}

func TestRWMutexWithWriter(t *testing.T) {
	var mutex RWMutex
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
	var mutex RWMutex
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
	var mutex RWMutex
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
	var mutex RWMutex
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
