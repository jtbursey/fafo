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
    ModeDiscovery WorkerMode = "Discovery"
    ModeFuzzy     WorkerMode = "Fuzzy"
    ModeAttack    WorkerMode = "Attack"
)

type Job struct {
    Mode     WorkerMode
    Action   string
    Priority int
    index    int
    Target   string
}

type JobQueue struct {
    queue    []Job
    mtx      *sync.Mutex
    inflight int
}

func (jq *JobQueue) Init() {
    jq.mtx = &sync.Mutex{}
    jq.mtx.Lock()
    defer jq.mtx.Unlock()
    jq.queue = make([]Job, 0)
    heap.Init(jq)
    jq.inflight = 0
}

func (jq *JobQueue) Push(job any) {
    jq.mtx.Lock()
    defer jq.mtx.Unlock()
    curJob := job.(Job)
    curJob.index = len(jq.queue)
    jq.queue = append(jq.queue, curJob)
}

func (jq *JobQueue) Pop() any {
    jq.mtx.Lock()
    defer jq.mtx.Unlock()
    if jq.Len() <= 0 {
        return nil
    }

    jq.inflight++
    n := len(jq.queue) - 1
    curJob := jq.queue[n]
    curJob.index = -1
    jq.queue = jq.queue[0:n]
    return curJob
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
    return len(jq.queue)
}

func (jq *JobQueue) Done() bool {
    jq.mtx.Lock()
    defer jq.mtx.Unlock()
    return jq.Len() == 0 && jq.inflight == 0
}

func (jq *JobQueue) Less(i, j int) bool {
    return jq.queue[i].Priority > jq.queue[j].Priority
}

func (jq *JobQueue) Swap(i, j int) {
    jq.queue[i], jq.queue[j] = jq.queue[j], jq.queue[i]
    jq.queue[i].index = i
    jq.queue[j].index = j
}
