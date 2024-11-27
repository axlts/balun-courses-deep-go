package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type orderable interface {
	numeric | ~string
}

type numeric interface {
	signed | unsigned | ~float32 | ~float64
}

type signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type node[K orderable, V any] struct {
	key K
	val V

	left, right *node[K, V]
}

type OrderedMap[K orderable, V any] struct {
	root *node[K, V]
	size int
}

func NewOrderedMap[K orderable, V any]() OrderedMap[K, V] {
	return OrderedMap[K, V]{}
}

func (m *OrderedMap[K, V]) Insert(key K, val V) {
	m.size++

	if m.root == nil {
		m.root = &node[K, V]{key: key, val: val}
		return
	}

	prv, cur := (*node[K, V])(nil), m.root
	for cur != nil {
		if cur.key == key {
			cur.val = val
			return
		}

		prv = cur
		if cur.key < key {
			cur = cur.right
		} else {
			cur = cur.left
		}
	}

	n := &node[K, V]{key: key, val: val}
	if prv.key > key {
		prv.left = n
	} else {
		prv.right = n
	}
}

func (m *OrderedMap[K, V]) Erase(key K) {
	if m.root == nil {
		return
	}

	prv, cur := (*node[K, V])(nil), m.root
	for cur != nil {
		if cur.key == key {
			break
		}

		prv = cur
		if cur.key < key {
			cur = cur.right
		} else {
			cur = cur.left
		}
	}
	if cur == nil {
		return
	}

	if cur.right == nil {
		if prv == nil {
			m.root = cur.left
		} else {
			if prv.left == cur {
				prv.left = cur
			} else {
				prv.right = cur.left
			}
		}
	} else {
		prv = (*node[K, V])(nil)
		lm := cur.right
		for lm.left != nil {
			prv, lm = lm, lm.left
		}
		if prv != nil {
			prv.left = lm.right
		} else {
			cur.right = lm.right
		}
		cur.key, cur.val = lm.key, lm.val
	}
	m.size--
}

func (m *OrderedMap[K, V]) Contains(key K) bool {
	cur := m.root
	for cur != nil {
		if cur.key == key {
			return true
		}
		if cur.key > key {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}
	return false
}

func (m *OrderedMap[K, V]) Size() int {
	return m.size
}

func (m *OrderedMap[K, V]) ForEach(action func(K, V)) {
	traverse(m.root, action)
}

func traverse[K orderable, V any](n *node[K, V], action func(K, V)) {
	if n == nil {
		return
	}
	traverse(n.left, action)
	action(n.key, n.val)
	traverse(n.right, action)
}

func TestCircularQueue(t *testing.T) {
	data := NewOrderedMap[int, int]()
	assert.Zero(t, data.Size())

	data.Insert(10, 10)
	data.Insert(5, 5)
	data.Insert(15, 15)
	data.Insert(2, 2)
	data.Insert(4, 4)
	data.Insert(12, 12)
	data.Insert(14, 14)

	assert.Equal(t, 7, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(3))
	assert.False(t, data.Contains(13))

	var keys []int
	expectedKeys := []int{2, 4, 5, 10, 12, 14, 15}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))

	data.Erase(15)
	data.Erase(14)
	data.Erase(2)

	assert.Equal(t, 4, data.Size())
	assert.True(t, data.Contains(4))
	assert.True(t, data.Contains(12))
	assert.False(t, data.Contains(2))
	assert.False(t, data.Contains(14))

	keys = nil
	expectedKeys = []int{4, 5, 10, 12}
	data.ForEach(func(key, _ int) {
		keys = append(keys, key)
	})

	assert.True(t, reflect.DeepEqual(expectedKeys, keys))
}
