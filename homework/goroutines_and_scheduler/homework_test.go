package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type heap struct {
	tasks []Task
	index map[int]int // task id -> index in array.
}

func newPriorityHeap() heap {
	return heap{
		tasks: []Task{},
		index: map[int]int{},
	}
}

func (h *heap) push(t Task) {
	h.tasks = append(h.tasks, t)

	if h.size() == 0 {
		h.index[t.Identifier] = 0
		return
	}

	i := h.size() - 1
	p := parent(i)
	for i > 0 && hasLowerPriority(h.tasks[p], h.tasks[i]) {
		h.swap(p, i)
		i = p
		p = parent(i)
	}
	h.index[t.Identifier] = i
}

func (h *heap) pop() Task {
	if len(h.tasks) == 0 {
		return Task{}
	}

	t := h.tasks[0]
	delete(h.index, t.Identifier)

	if len(h.tasks) == 1 {
		h.tasks = h.tasks[:0]
		return t
	}

	h.remove(0)
	return t
}

func (h *heap) update(id int, prio int) {
	i, ok := h.index[id]
	if !ok {
		return
	}

	t := h.tasks[i]
	t.Priority = prio

	h.remove(i)
	h.push(t)
}

func (h *heap) remove(i int) {
	h.swap(i, h.size()-1)
	h.tasks = h.tasks[:h.size()-1]
	h.heapify(i)
}

func (h *heap) heapify(i int) {
	for {
		l := left(i)
		r := right(i)

		lgst := i
		if l < h.size() && hasLowerPriority(h.tasks[lgst], h.tasks[l]) {
			lgst = l
		}
		if r < h.size() && hasLowerPriority(h.tasks[lgst], h.tasks[r]) {
			lgst = r
		}

		if lgst == i {
			break
		}

		h.swap(lgst, i)
		i = lgst
	}
}

func (h *heap) swap(i, j int) {
	h.index[h.tasks[i].Identifier] = j
	h.index[h.tasks[j].Identifier] = i

	h.tasks[i], h.tasks[j] = h.tasks[j], h.tasks[i]
}

func (h *heap) size() int {
	return len(h.tasks)
}

func parent(i int) int {
	return (i - 1) / 2
}

func left(i int) int {
	return i*2 + 1
}

func right(i int) int {
	return i*2 + 2
}

type Task struct {
	Identifier int
	Priority   int
}

func hasLowerPriority(lhs, rhs Task) bool {
	return lhs.Priority < rhs.Priority
}

type Scheduler struct {
	h heap
}

func NewScheduler() Scheduler {
	return Scheduler{h: newPriorityHeap()}
}

func (s *Scheduler) AddTask(t Task) {
	s.h.push(t)
}

func (s *Scheduler) ChangeTaskPriority(id int, prio int) {
	s.h.update(id, prio)
}

func (s *Scheduler) GetTask() Task {
	return s.h.pop()
}

func TestTrace(t *testing.T) {
	task1 := Task{Identifier: 1, Priority: 10}
	task2 := Task{Identifier: 2, Priority: 20}
	task3 := Task{Identifier: 3, Priority: 30}
	task4 := Task{Identifier: 4, Priority: 40}
	task5 := Task{Identifier: 5, Priority: 50}

	scheduler := NewScheduler()
	scheduler.AddTask(task1)
	scheduler.AddTask(task2)
	scheduler.AddTask(task3)
	scheduler.AddTask(task4)
	scheduler.AddTask(task5)

	task := scheduler.GetTask()
	assert.Equal(t, task5, task)

	task = scheduler.GetTask()
	assert.Equal(t, task4, task)

	task1.Priority = 100 // fix test.
	scheduler.ChangeTaskPriority(1, 100)

	task = scheduler.GetTask()
	assert.Equal(t, task1, task)

	task = scheduler.GetTask()
	assert.Equal(t, task3, task)
}
