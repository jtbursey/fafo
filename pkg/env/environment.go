// Joseph Bursey <jbursey@tevora.com>

package env

import (
	"sync"

	"fafo/pkg/fact"
	"fafo/pkg/httpclient"
	"fafo/pkg/job"
)

type Env struct {
	Jobqueue  job.JobQueue					// The queue of jobs for workers to pull from
	Cfg       Config						// Extra Config (sommeday to be set by the user)
	targets   map[string]fact.Target		// Keep inforation about known targets
	targetmtx sync.Mutex
	Client    httpclient.HttpClient
	CorpusCh  chan string					// Channel for pushing corpus updates (to mgr)
	JobCh     chan job.Job					// Channel for pushing more jobs (to mgr)
	FactCh    chan fact.Fact				// Channel for pushing facts/results  (to mgr)
}

func (env *Env) PullTarget(name string) *fact.Target {
	env.targetmtx.Lock()
	defer env.targetmtx.Unlock()
	if env.targets == nil {
		return nil
	}

	ret, ok := env.targets[name]
	if ok {
		return &ret
	}
	return nil
}

// Lock must already be held!
func (env *Env) mergeTarget(new fact.Target) {
	old := env.targets[new.Url]

	if old.IsAlive <= 0 && new.IsAlive >= 0 {
		old.IsAlive = new.IsAlive
	}
}

func (env *Env) PushTarget(target fact.Target) {
	env.targetmtx.Lock()
	defer env.targetmtx.Unlock()
	if env.targets == nil {
		env.targets = make(map[string]fact.Target)
	}

	_, ok := env.targets[target.Url]
	if ok {
		env.mergeTarget(target)
	} else {
		env.targets[target.Url] = target
	}
	
}
