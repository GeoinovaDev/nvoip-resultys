package queuetime

import (
	"sync"
	"time"
)

type QueueTime struct {
	mx       *sync.Mutex
	Interval int

	list []func()
}

func New(interval int) *QueueTime {
	return &QueueTime{
		mx:       &sync.Mutex{},
		list:     []func(){},
		Interval: interval,
	}
}

func (q *QueueTime) Run() {
	go q._run()
}

func (q *QueueTime) Push(fn func()) {
	q.mx.Lock()
	defer q.mx.Unlock()

	q.list = append(q.list, fn)
}

func (q *QueueTime) Pop() (func(), bool) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if len(q.list) == 0 {
		return nil, false
	}

	firstItem := q.list[0]
	q.list = append(q.list[:0], q.list[1:]...)

	return firstItem, true
}

func (q *QueueTime) _run() {
	for {
		time.Sleep(time.Duration(q.Interval) * time.Millisecond)
		fn, called := q.Pop()
		if called {
			fn()
		}
	}
}
