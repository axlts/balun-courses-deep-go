package main

import "unsafe"

const intlen = int(unsafe.Sizeof(*(new(int))))

type array struct {
	len int
	ptr unsafe.Pointer
}

func alloc(size int) *array {
	return &array{ptr: unsafe.Pointer(&(make([]int, size))[0]), len: size}
}

func (a *array) get(idx int) int {
	return *(*int)(unsafe.Add(a.ptr, idx*intlen))
}

func (a *array) set(idx int, val int) {
	*(*int)(unsafe.Add(a.ptr, idx*intlen)) = val
}

func (a *array) slice() []int {
	return unsafe.Slice((*int)(a.ptr), a.len)
}
