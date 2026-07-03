// Joseph Bursey <jbursey@tevora.com>

// This file defines various discovery actions that can be carried out by workers

package worker

import (
	"slices"

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

// Check if the target is alive
func (w *Worker) CheckAlive(target *fact.Target, env *env.Env) {
	resp := env.Client.Get(target.Url)

	res := fact.Fact{
		Url:   target.Url,
		Port:  target.Port,
		Novel: make(map[fact.FactKey]fact.FactValue),
	}

	if resp == nil {
		w.Errf("Failed to GET target: %v\n", target.Url)
		res.Novel[fact.IsAlive] = fact.FactFalse
		//env.FactCh <- res // Don't send the fact to keep down on memory
		return
	}

	if slices.Contains(aliveValid, resp.StatusCode) {
		w.Logf(0, "%v\n", pretty.Response(resp, target.Url))
		res.Novel[fact.IsAlive] = fact.FactTrue
		res.Novel[fact.Exists] = fact.FactTrue
		env.FactCh <- res
		

		if target.Type == fact.TargetDomain || target.Type == fact.TargetPath {
			// TODO: Push a fuzz directories job
			// TODO: Push a fuzz files job
		}

		// TODO: Push a fuzzy job


	} else {
		w.Logf(2, "%v\n", pretty.Response(resp, target.Url))
		res.Novel[fact.IsAlive] = fact.FactTrue
		res.Novel[fact.Exists] = fact.FactFalse
		env.FactCh <- res
	}
}

// Match target to known fingerprints
func (w *Worker) FingerPrint() {

}

func (w *Worker) FuzzCommonPorts() {

}

// Path fuzzing
func (w *Worker) FuzzDirectories() {

}

func (w *Worker) FuzzFiles() {
	
}

// WP Scan
func (w *Worker) WordPressScan() {

}

func (w *Worker) discoveryDispatch(j *job.Job, t *fact.Target, e *env.Env) {
	if j.Action == ActionCheckAlive {
		w.CheckAlive(t, e)
	}
}
