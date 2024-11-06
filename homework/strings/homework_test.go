package main

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type COWBuffer struct {
	data []byte
	refs *int

	closed bool
}

func NewCOWBuffer(data []byte) *COWBuffer {
	cow := &COWBuffer{
		data: data,
		refs: ptr(1),
	}
	runtime.SetFinalizer(cow, (*COWBuffer).Close)
	return cow
}

func (b *COWBuffer) Clone() *COWBuffer {
	if b.closed {
		return nil
	}

	*b.refs++
	cow := &COWBuffer{
		data: b.data,
		refs: b.refs,
	}
	runtime.SetFinalizer(cow, (*COWBuffer).Close)
	return cow
}

func (b *COWBuffer) Close() {
	if b.closed {
		return
	}

	*b.refs--
	b.closed = true
}

func (b *COWBuffer) Update(index int, value byte) bool {
	if b.closed || index < 0 || index >= len(b.data) {
		return false
	}

	if *b.refs > 1 {
		*b.refs--

		cpy := make([]byte, len(b.data))
		copy(cpy, b.data)
		b.data = cpy
		b.refs = ptr(1)
	}

	*(*byte)(unsafe.Add(unsafe.Pointer(&b.data[0]), index*(1<<3))) = value
	return true
}

func (b *COWBuffer) String() string {
	if b.closed {
		return ""
	}
	return unsafe.String(unsafe.SliceData(b.data), len(b.data))
}

func ptr[T any](x T) *T {
	return &x
}

func TestCOWBuffer(t *testing.T) {
	data := []byte{'a', 'b', 'c', 'd'}
	buffer := NewCOWBuffer(data)
	defer buffer.Close()

	copy1 := buffer.Clone()
	copy2 := buffer.Clone()

	assert.Equal(t, unsafe.SliceData(data), unsafe.SliceData(buffer.data))
	assert.Equal(t, unsafe.SliceData(buffer.data), unsafe.SliceData(copy1.data))
	assert.Equal(t, unsafe.SliceData(copy1.data), unsafe.SliceData(copy2.data))

	assert.True(t, (*byte)(unsafe.SliceData(data)) == unsafe.StringData(buffer.String()))
	assert.True(t, (*byte)(unsafe.StringData(buffer.String())) == unsafe.StringData(copy1.String()))
	assert.True(t, (*byte)(unsafe.StringData(copy1.String())) == unsafe.StringData(copy2.String()))

	assert.True(t, buffer.Update(0, 'g'))
	assert.False(t, buffer.Update(-1, 'g'))
	assert.False(t, buffer.Update(4, 'g'))

	assert.True(t, reflect.DeepEqual([]byte{'g', 'b', 'c', 'd'}, buffer.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy1.data))
	assert.True(t, reflect.DeepEqual([]byte{'a', 'b', 'c', 'd'}, copy2.data))

	assert.NotEqual(t, unsafe.SliceData(buffer.data), unsafe.SliceData(copy1.data))
	assert.Equal(t, unsafe.SliceData(copy1.data), unsafe.SliceData(copy2.data))

	copy1.Close()

	previous := copy2.data
	copy2.Update(0, 'f')
	current := copy2.data

	// 1 reference - don't need to copy buffer during update
	assert.Equal(t, unsafe.SliceData(previous), unsafe.SliceData(current))

	copy2.Close()
}
