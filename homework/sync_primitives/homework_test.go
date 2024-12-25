package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type RWMutex struct {
	inmtx, outmtx sync.Mutex
	rcnt, wcnt    int
}

func (m *RWMutex) Lock() {
	m.inmtx.Lock()
	defer m.inmtx.Unlock()

	m.wcnt++
	m.outmtx.Lock()
}

func (m *RWMutex) Unlock() {
	m.inmtx.Lock()
	defer m.inmtx.Unlock()

	m.wcnt--
	m.outmtx.Unlock()
}

func (m *RWMutex) RLock() {
	m.inmtx.Lock()
	defer m.inmtx.Unlock()

	if m.rcnt == 0 || m.wcnt > 0 {
		m.outmtx.Lock()
	}
	m.rcnt++
}

func (m *RWMutex) RUnlock() {
	m.inmtx.Lock()
	defer m.inmtx.Unlock()

	if m.rcnt == 0 {
		m.outmtx.Unlock()
	}
	m.rcnt--
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
