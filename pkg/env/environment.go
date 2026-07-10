// Joseph Bursey <jbursey@tevora.com>

package env

import (
    "fmt"
    "net/http"

    "fafo/pkg/fact"
    "fafo/pkg/httpclient"
    "fafo/pkg/job"
    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

type Env struct {
    Jobqueue    job.JobQueue                  // The queue of jobs for workers to pull from
    Cfg         Config                        // Extra Config (sommeday to be set by the user)
    FirstTarget fact.Target
    Targets     fact.TargetMap                // Keep information about known targets
    Client      httpclient.HttpClient
    ScrShCh     chan http.Request             // Channel for pushing screenshot requests
    JobCh       chan job.Job                  // Channel for pushing more jobs (to mgr)
    FactCh      chan fact.Target              // Channel for pushing facts/results  (to mgr)
}

func (e *Env) Debug() {
    log.Logf(0, "%v\n", pretty.Config("Target", fmt.Sprintf("%v (%v)", e.FirstTarget.Url, e.FirstTarget.Type)))
    e.Cfg.Debug()
    log.Log(0, "\n\n")
}
