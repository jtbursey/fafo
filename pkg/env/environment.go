// Joseph Bursey <jbursey@tevora.com>

package env

import (
    "fafo/pkg/fact"
    "fafo/pkg/httpclient"
    "fafo/pkg/job"
)

type Env struct {
    Jobqueue  job.JobQueue                  // The queue of jobs for workers to pull from
    Cfg       Config                        // Extra Config (sommeday to be set by the user)
    Targets   fact.TargetMap                // Keep information about known targets
    Client    httpclient.HttpClient
    ScrShCh   chan string                   // Channel for pushing screenshot requests
    JobCh     chan job.Job                  // Channel for pushing more jobs (to mgr)
    FactCh    chan fact.Target              // Channel for pushing facts/results  (to mgr)
}
