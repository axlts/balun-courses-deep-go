package main

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type unsigned interface {
	uint8 | uint16 | uint32 | uint64
}

func ToLittleEndianBitOps[T unsigned](be T) (le T) {
	if be == 0 {
		return 0
	}

	size := T(unsafe.Sizeof(be))

	var i T
	for ; i < size; i++ {
		le <<= 1 << 3
		le |= be & 0xff
		be >>= 1 << 3
	}
	return
}

func ToLittleEndianUnsafe[T unsigned](be T) (le T) {
	if be == 0 {
		return 0
	}

	size := unsafe.Sizeof(be)

	buf := make([]byte, size)
	for i := len(buf) - 1; i >= 0; i-- {
		buf[i] = byte(be & 0xff)
		be >>= 1 << 3
	}
	return *(*T)(unsafe.Pointer(&buf[0]))
}

func TestSerializationProperties(t *testing.T) {
	tests := map[string]struct {
		number uint32
		result uint32
	}{
		"test case #1": {
			number: 0x00000000,
			result: 0x00000000,
		},
		"test case #2": {
			number: 0xFFFFFFFF,
			result: 0xFFFFFFFF,
		},
		"test case #3": {
			number: 0x00FF00FF,
			result: 0xFF00FF00,
		},
		"test case #4": {
			number: 0x0000FFFF,
			result: 0xFFFF0000,
		},
		"test case #5": {
			number: 0x01020304,
			result: 0x04030201,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := ToLittleEndianBitOps[uint32](test.number)
			assert.Equal(t, test.result, result)

			result = ToLittleEndianUnsafe[uint32](test.number)
			assert.Equal(t, test.result, result)
		})
	}
}

func BenchmarkToLittleEndianBitOps(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ToLittleEndianBitOps[uint32](0x01020304)
	}
}

func BenchmarkToLittleEndianUnsafe(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ToLittleEndianUnsafe[uint32](0x01020304)
	}
}
