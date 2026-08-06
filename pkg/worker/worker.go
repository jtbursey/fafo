// Joseph Bursey  <jbursey@tevora.com>

package worker

import (
    "context"
    "fmt"
    "net/url"

    "fafo/pkg/env"
    "fafo/pkg/fact"
    "fafo/pkg/fam"
    "fafo/pkg/job"
    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

type WorkerStatus string

const (
    StatusStartup WorkerStatus = "startup"
    StatusIdle    WorkerStatus = "idle"
    StatusWorking WorkerStatus = "working"

    VError  int = log.VError
    VWarn   int = log.VError
    VMode   int = log.V3
    VStatus int = log.V7
)

type Worker struct {
    ctx    context.Context
    id     uint
    status WorkerStatus
    mode   job.WorkerMode
}

func (w *Worker) IdString() string {
    return fmt.Sprintf("Worker %v", w.id)
}

func (w *Worker) prefix() string {
    if log.Verb(log.VPrefix) {
        return fmt.Sprintf("%*s", pretty.PrefixWidth, fmt.Sprintf("[%v]: ", w.IdString()))
    }
    return ""
}

func (w *Worker) Logf(v int, msg string, args ...any) {
    log.Logf(v, w.prefix()+msg, args...)
}

func (w *Worker) Log(v int, msg string) {
    w.Logf(v, "%v", msg)
}

func (w *Worker) Errf(msg string, args ...any) {
    log.Logf(VError, fmt.Sprintf("%v%v: %v", w.prefix(), pretty.Orange("Error"), msg), args...)
}

func (w *Worker) Warnf(msg string, args ...any) {
    log.Logf(VWarn, fmt.Sprintf("%v%v: %v", w.prefix(), pretty.Orange("Warning"), msg), args...)
}

func (w *Worker) newStatus(status WorkerStatus) {
    w.status = status
    w.Logf(VStatus, "New status: %v\n", w.status)
}

func (w *Worker) newMode(mode job.WorkerMode) {
    w.mode = mode
    w.Logf(VMode, "Switching to %v mode\n", w.mode)
}

func (w *Worker) resetMode() {
    w.mode = job.ModeNone
}

func (w *Worker) dispatch(job *job.Job, t *fact.Target, env *env.Env) {
    if action, ok := env.Actions[job.Action]; ok {
        w.newMode(action.Mode)
        f := fam.Fam{
            Caller: w.IdString(),
        }
        f.Run(t, &action, env)
    } else {
        w.Warnf("Found unimplemented Job Action: %v\n", job.Action)
    }
    w.resetMode()
}

func (w *Worker) checkDone() bool {
    select {
        case <-w.ctx.Done():
            return true
        default:
            return false
        }
}

func (w *Worker) Loop(id uint, env *env.Env) {
    for {
        // Right now  this is enough.
        // The only time we call cancel is when Manager exits the loop cleanly, meaning no work is in flight
        if w.checkDone() {
            return
        }

        if maybeJob := env.Jobqueue.Pop(); maybeJob != nil {
            curJob := maybeJob.(job.Job)
            w.newStatus(StatusWorking)

            // TODO: this is going to be hyper broken once we start doing variable/argument fuzzing
                // The fix will be to only work on the url without argumments (host+path?)
                // Should jobs get pushed with arguments? or is that up to the new actions to decide?
            target := env.Targets.Pull(curJob.Target)
            if target == nil {
                target = &fact.Target{
                    Facts: make(map[fact.FactKey][]string),
                }
                if tmpUrl, err := url.Parse(fact.ChopSlash(curJob.Target)); err != nil {
                    w.Errf("Failed to parse Url from Job: %v: %v\n", curJob.Target, err)
                    w.newStatus(StatusIdle)
                    env.Jobqueue.Finish()
                    continue
                } else {
                    target.Url = tmpUrl
                }
            }

            w.dispatch(&curJob, target, env)
            w.newStatus(StatusIdle)
            env.Jobqueue.Finish()
        }
    }
}

func Run(ctx context.Context, i uint, env *env.Env) {
    w := &Worker{
        ctx:    ctx,
        id:     i,
        status: StatusStartup,
    }

    w.newStatus(StatusIdle)
    w.Loop(i, env)
}
