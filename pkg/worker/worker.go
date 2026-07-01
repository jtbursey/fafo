// Joseph Bursey  <jbursey@tevora.com>

package worker

import (
	"fafo/pkg/env"
	//"fafo/pkg/fact"
	//"fafo/pkg/job"
	"fafo/pkg/log"
)

type WorkerStatus string

const (
	StatusStartup WorkerStatus = "startup"
	StatusIdle    WorkerStatus = "idle"
	StatusWorking WorkerStatus = "working"
)

type Worker struct {
	id     uint
	status WorkerStatus
}

func (w Worker) Log(v int, msg string) {
	w.Logf(v, "%v", msg)
}

func (w Worker) Logf(v int, msg string, args ...any) {
	log.Logf(v, "[Worker %v]: "+msg, append([]any{w.id}, args...)...)
}

func (w *Worker) newStatus(status WorkerStatus) {
	w.status = status
	w.Logf(1, "New status: %v\n", w.status)
}

func (w *Worker) Dispatch() {

}

func (w *Worker) Loop(id uint, env *env.Env) {
	for ;; {
		if env.Jobqueue.Poll() {
			w.newStatus(StatusWorking)
			curJob := env.Jobqueue.Pop()
			w.Logf(1, "Switching to %v mode for %v\n", curJob.Mode, curJob.Target)
			
			// Pull the corpus

			target := env.PullTarget(curJob.Target)
			if target == nil {
				w.Logf(0, "Error pulling target: %v. Target does not exist.\n", curJob.Target)
				w.newStatus(StatusIdle)
				continue
			}
			
			// Carry out the job
			w.CheckAlive(target, *env)

			// Push new jobs
			// push corpus sync
			env.PushTarget(*target)
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
