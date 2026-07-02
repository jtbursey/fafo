// Joseph Bursey <jbursey@tevora.com>

package job

import (
	"container/heap"
	"sync"

	"fafo/pkg/log"
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
	inflight int
}

func (jq *JobQueue) Init() {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	jq.queue = make(PriorityQueue, 0)
	heap.Init(&jq.queue)
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
	jq.inflight++
	return jq.queue.Pop().(*Job)
}

func (jq *JobQueue) Finish() {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	if jq.inflight == 0 {
		log.Err("Inflight counter decrimented when no jobs were in flight!\n")
		return
	}
	jq.inflight--
}

func (jq *JobQueue) Len() int {
	return jq.queue.Len()
}

func (jq *JobQueue) Poll() bool {
	return jq.Len() > 0
}

func (jq *JobQueue) Done() bool {
	jq.mtx.Lock()
	defer jq.mtx.Unlock()
	return jq.Len() == 0 && jq.inflight == 0
}
