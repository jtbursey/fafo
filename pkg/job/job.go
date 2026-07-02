// Joseph Bursey <jbursey@tevora.com>

package job

import (
	"container/heap"
	"sync"
)

type WorkerMode string

const (
	ModeNone      WorkerMode = "none"
	ModeDiscovery WorkerMode = "discovery"
	ModeFuzzy     WorkerMode = "fuzzy"
	ModeAttack    WorkerMode = "attack"
)

type Action string

type Job struct {
	Mode     WorkerMode
	Action   Action
	Priority int
	index    int
	Target   string
}

type JobQueue struct {
	queue    PriorityQueue
	mtx      sync.Mutex
	waiting  int
	inflight int
}

func (jq *JobQueue) Init() {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	jq.queue = make(PriorityQueue, 0)
	heap.Init(&jq.queue)
	jq.waiting = 0
	jq.inflight = 0
}

func (jq *JobQueue) Push(job *Job) {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	jq.queue.Push(job)
}

func (jq *JobQueue) Pop() *Job {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	return jq.queue.Pop().(*Job)
}

func (jq JobQueue) Len() int {
	return jq.queue.Len()
}

func (jq JobQueue) Poll() bool {
	return jq.Len() > 0
}
