// Joseph Bursey <jbursey@tevora.com>

package job

import (
	"container/heap"
	"sync"
)

type WorkerMode string

const (
	ModeDiscovery WorkerMode = "discovery"
	ModeFuzzy     WorkerMode = "fuzzy"
	ModeAttack    WorkerMode = "attack"
)

type Job struct {
	Mode     WorkerMode
	Priority int
	index    int
	Target   string
}

type JobQueue struct {
	queue PriorityQueue
	mtx   sync.Mutex
}

func (jq *JobQueue) Init() {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	jq.queue = make(PriorityQueue, 0)
	heap.Init(&jq.queue)
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
