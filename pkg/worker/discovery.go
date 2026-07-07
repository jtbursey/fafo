// Joseph Bursey <jbursey@tevora.com>

// This file defines various discovery actions that can be carried out by workers

package worker

import (
	"net/http"
	"slices"
	"sync"

	"fafo/pkg/env"
	"fafo/pkg/fact"
	"fafo/pkg/pretty"
	"fafo/pkg/job"
)

const (
	ActionCheckAlive      job.Action = "CheckAlive"
	ActionFuzzDirectories job.Action = "FuzzDirectories"
	ActionFuzzFiles       job.Action = "FuzzFiles"
	ActionWordPressScan   job.Action = "WordPressScan"
)

var (
	// Defaults from ffuf
	aliveValid = []int {200, 204, 301, 302, 307, 401}
)

func (w *Worker) getAndClose(target *fact.Target, env *env.Env) *http.Response {
	resp := env.Client.Get(target.Url)
	if resp == nil {
		w.Errf("Failed to GET target: %v\n", target.Url)
		return nil
	}
	env.Client.DropBody(resp)
	return resp
}

// Check if the target is alive
func (w *Worker) CheckAlive(target *fact.Target, env *env.Env) {
	res := fact.Target{
		Url:   target.Url,
		Type:  target.Type,
		Port:  target.Port,
		Facts: make(map[fact.FactKey]fact.FactValue),
	}
	
	resp := w.getAndClose(target, env)
	if resp == nil {
		res.Facts[fact.IsAlive] = fact.False
		//env.FactCh <- res // Don't send the fact to keep down on memory
		return
	}

	if slices.Contains(aliveValid, resp.StatusCode) {
		w.Logf(0, "%v\n", pretty.Response(resp, target.Url))
		res.Facts[fact.IsAlive] = fact.True
		res.Facts[fact.Exists] = fact.True
		env.FactCh <- res

		if target.Type == fact.TargetDomain || target.Type == fact.TargetPath {
			env.JobCh <- job.Job{
				Mode:     job.ModeDiscovery,
				Action:   ActionFuzzDirectories,
				Priority: 5,
				Target:   target.Key(),
			}

			env.JobCh <- job.Job{
				Mode:     job.ModeDiscovery,
				Action:   ActionFuzzFiles,
				Priority: 5,
				Target:   target.Key(),
			}
		}

		// TODO: Push a fuzzy job


	} else {
		w.Logf(2, "%v\n", pretty.Response(resp, target.Url))
		res.Facts[fact.IsAlive] = fact.True
		res.Facts[fact.Exists] = fact.False
		env.FactCh <- res
	}
}

// Match target to known fingerprints
func (w *Worker) FingerPrint() {

}

func (w *Worker) FuzzCommonPorts() {

}

func (w *Worker) fuzzFromList(target *fact.Target, env *env.Env, listFile string) {
	w.Logf(10, "Fuzzing %v with %v\n", target.Url, listFile)
	
	// Take hint from MaxCalls and spawn that many workers.
	var wg sync.WaitGroup
	ch := make(chan string, env.Cfg.ClientCfg.MaxCalls*2)
	done := false
	wg.Go(func() {w.channelFile(listFile, ch, &done)})
	for i := 0; i < env.Cfg.ClientCfg.MaxCalls; i++ {
		wg.Go(func() {
			for {
				select {
				case item := <- ch:
					res := fact.Target{
						Url:   fact.UrlAppend(target.Url, item),
						Type:  fact.TargetPath,
						Port:  target.Port,
						Facts: make(map[fact.FactKey]fact.FactValue),
					}

					// TODO: pull the target and make sure this is not redundant

					resp := w.getAndClose(&res, env)
					if resp == nil {
						res.Facts[fact.IsAlive] = fact.False
						//env.FactCh <- res // Don't send the fact to keep down on memory
						break
					}

					if slices.Contains(aliveValid, resp.StatusCode) {
						w.Logf(0, "%v\n", pretty.Response(resp, res.Url))
						res.Facts[fact.Exists] = fact.True
						env.FactCh <- res

						// TODO: push a job?
					} else {
						w.Logf(2, "%v\n", pretty.Response(resp, res.Url))
						// JTBursey: No need to push these if they are not found
						//res.Facts[fact.Exists] = fact.False
						//env.FactCh <- res
					}
				default:
					if done { return }
				}
			}
		})
	}

	wg.Wait()
}

// Path fuzzing
func (w *Worker) FuzzDirectories(target *fact.Target, env *env.Env) {
	listFile := fact.UrlAppend(env.Cfg.Seclists, env.Cfg.FuzzDirList)

	w.fuzzFromList(target, env, listFile)
}

func (w *Worker) FuzzFiles(target *fact.Target, env *env.Env) {
	listFile := fact.UrlAppend(env.Cfg.Seclists, env.Cfg.FuzzFileList)

	w.fuzzFromList(target, env, listFile)
}

// WP Scan
func (w *Worker) WordPressScan() {

}

func (w *Worker) discoveryDispatch(j *job.Job, t *fact.Target, e *env.Env) {
	switch {
	case j.Action == ActionCheckAlive:
		w.CheckAlive(t, e)
	case j.Action == ActionFuzzDirectories:
		w.FuzzDirectories(t, e)
	case j.Action == ActionFuzzFiles:
		w.FuzzFiles(t, e)
	}
}
