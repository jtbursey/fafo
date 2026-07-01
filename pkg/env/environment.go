// Joseph Bursey <jbursey@tevora.com>

package env

import (
	//"fafo/pkg/corpus"
	"fafo/pkg/job"
)

type Env struct {
	Jobqueue job.JobQueue		// The queue of jobs for workers to pull from
	Cfg      Config				// Extra Config (sommeday to be set by the user)
	CorpusCh chan string		// Channel for pushing corpus updates (to mgr)
	JobCh    chan job.Job		// Channel for pushing more jobs (to mgr)
	FactCh   chan string		// Channel for pushing facts/results  (to mgr)
}
