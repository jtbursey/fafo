// Joseph Bursey  <jbursey@tevora.com>

package worker

import (
	"time"

	"fafo/pkg/env"
	"fafo/pkg/log"
)

const (
	StatusStartup string = "startup"
	StatusIdle    string = "idle"
	StatusWorking string = "working"
)

type Worker struct {
	id     uint
	status string
}

func (w Worker) Log(v int, msg string) {
	w.Logf(v, "%v", msg)
}

func (w Worker) Logf(v int, msg string, args ...any) {
	log.Logf(v, "[Worker %v]: "+msg, append([]any{w.id}, args...)...)
}

func (w *Worker) newStatus(status string) {
	w.status = status
	w.Logf(1, "New status: %v\n", w.status)
}

func (w *Worker) Loop(id uint, env *env.Env) {
	for ;; {
		w.Log(0, "I exist!\n")
		time.Sleep(2 * time.Second)

		if env.Jobqueue.Len() > 0 {
			w.newStatus(StatusWorking)
			curJob := env.Jobqueue.Pop()
			w.Logf(1, "Beginning %v on %v\n", curJob.Mode, curJob.Target)
			// Poll from Priority Queue for jobs
			// Pull the corpus
			// Carry out the job
			// Push new jobs
			// push corpus sync
			time.Sleep(2 * time.Second)
			env.FactCh <- "hello\n"
			w.newStatus(StatusIdle)
		}
	}
}

func Run(i uint, env *env.Env) {
	w := &Worker{
		id:     i,
		status: StatusStartup,
	}

	w.newStatus(StatusIdle)
	w.Loop(i, env)
}
