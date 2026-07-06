// Joseph Bursey  <jbursey@tevora.com>

package worker

import (
	"fmt"

	"fafo/pkg/env"
	"fafo/pkg/fact"
	"fafo/pkg/job"
	"fafo/pkg/log"
	"fafo/pkg/pretty"
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
	mode   job.WorkerMode
}

func (w *Worker) Logf(v int, msg string, args ...any) {
	if log.Verb(v) {
		log.Logf(3, "%-13v", fmt.Sprintf("[Worker %v]: ", w.id))
		log.Logf(v, msg, args...)
	}
}

func (w *Worker) Log(v int, msg string) {
	w.Logf(v, "%v", msg)
}

func (w *Worker) Errf(msg string, args ...any) {
	log.Logf(0, fmt.Sprintf("%-13v%v: %v", fmt.Sprintf("[Worker %v]: ", w.id), pretty.Orange("Error"), msg), args...)
}

func (w *Worker) newStatus(status WorkerStatus) {
	w.status = status
	w.Logf(3, "New status: %v\n", w.status)
}

func (w *Worker) newMode(mode job.WorkerMode) {
	w.mode = mode
	w.Logf(3, "Switching to %v mode\n", w.mode)
}

func (w *Worker) resetMode() {
	w.mode = job.ModeNone
}

func (w *Worker) dispatch(j *job.Job, t *fact.Target, e *env.Env) {
	w.newMode(j.Mode)
	if j.Mode == job.ModeDiscovery {
		w.discoveryDispatch(j, t, e)
	}
	w.resetMode()
}

func (w *Worker) Loop(id uint, env *env.Env) {
	for ;; {
		if env.Jobqueue.Poll() {
			w.newStatus(StatusWorking)
			curJob := env.Jobqueue.Pop()
			
			// Pull the corpus

			target := env.Targets.Pull(curJob.Target)
			if target == nil {
				w.Errf("Failed to pull target: %v: Target does not exist.\n", curJob.Target)
				w.newStatus(StatusIdle)
				env.Jobqueue.Finish()
				continue
			}

			w.dispatch(curJob, target, env)

			// Push new jobs
			// push corpus sync
			
			w.newStatus(StatusIdle)
			env.Jobqueue.Finish()
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
