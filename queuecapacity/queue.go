package queuecapacity

import (
	"sync"
	"time"
)

const (
	QUEUED  = 0
	RUNNING = 1
	DONE    = 2
)

// QueueItem ...
type QueueItem struct {
	ID      int
	status  int
	Context interface{}
}

// QueueCapacity ...
type QueueCapacity struct {
	items          []*QueueItem
	mx             *sync.Mutex
	max            int
	workerInterval int
	counter        int

	fnPush func(*QueueItem)
}

// New ...
func New(capacity int) *QueueCapacity {
	return &QueueCapacity{
		items:          []*QueueItem{},
		mx:             &sync.Mutex{},
		max:            capacity,
		workerInterval: 10,
		counter:        0,
		fnPush:         nil,
	}
}

// SetWorkerInterval ...
func (q *QueueCapacity) SetWorkerInterval(time int) {
	q.workerInterval = time
}

// SetConcurrency ...
func (q *QueueCapacity) SetConcurrency(max int) {
	q.max = max
}

// OnPush ...
func (q *QueueCapacity) OnPush(fn func(*QueueItem)) {
	q.fnPush = fn
}

// AddItem ...
func (q *QueueCapacity) AddItem(context interface{}) *QueueItem {
	q.mx.Lock()
	defer q.mx.Unlock()

	q.counter++
	item := &QueueItem{
		ID:      q.counter,
		status:  QUEUED,
		Context: context,
	}
	q.items = append(q.items, item)

	return item
}

// RemoveItem ...
func (q *QueueCapacity) RemoveItem(id int) {
	q.mx.Lock()
	defer q.mx.Unlock()

	for i := 0; i < len(q.items); i++ {
		if q.items[i].ID == id {
			q.items = append(q.items[:i], q.items[i+1:]...)
			break
		}
	}
}

// Run ...
func (q *QueueCapacity) Run() {
	go q.worker()
}

func (q *QueueCapacity) fetchItemByStatus(status int) *QueueItem {
	for i := 0; i < len(q.items); i++ {
		if q.items[i].status == status {
			return q.items[i]
		}
	}

	return nil
}

func (q *QueueCapacity) countItemByStatus(status int) int {
	counter := 0

	for i := 0; i < len(q.items); i++ {
		if q.items[i].status == status {
			counter++
		}
	}

	return counter
}

func (q *QueueCapacity) worker() {
	for {
		var item *QueueItem = nil

		q.mx.Lock()
		countItemRunning := q.countItemByStatus(RUNNING)
		if countItemRunning < q.max {
			item = q.fetchItemByStatus(QUEUED)
			if item != nil && q.fnPush != nil {
				item.status = RUNNING
				go q.fnPush(item)
			}
		}
		q.mx.Unlock()

		time.Sleep(time.Duration(q.workerInterval) * time.Millisecond)
	}
}
