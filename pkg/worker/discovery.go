// Joseph Bursey <jbursey@tevora.com>

// This file defines various discovery actions that can be carried out by workers

package worker

import (
	//"bufio"
	"fmt"
	"os"
	"slices"
	"strings"

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
	env.Client.DropBody(resp)

	if slices.Contains(aliveValid, resp.StatusCode) {
		w.Logf(0, "%v\n", pretty.Response(resp, target.Url))
		res.Novel[fact.IsAlive] = fact.FactTrue
		res.Novel[fact.Exists] = fact.FactTrue
		env.FactCh <- res
		

		if target.Type == fact.TargetDomain || target.Type == fact.TargetPath {
			env.JobCh <- job.Job{
				Mode:     job.ModeDiscovery,
				Action:   ActionFuzzDirectories,
				Priority: 5,
				Target:   target.Key(),
			}

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

func (w *Worker) fuzzFromList(target *fact.Target, env *env.Env, listFile string) {
	w.Logf(10, "Fuzzing with list from %v\n", listFile)
	file, err := os.Open(listFile)
	if err != nil {
		w.Errf("Failed to open file %v.", listFile)
		return
	}

	file.Close()
}

// Path fuzzing
func (w *Worker) FuzzDirectories(target *fact.Target, env *env.Env) {
	sep := ""
	if len(env.Cfg.Seclists) > 0 && !strings.HasSuffix(env.Cfg.Seclists, "/") {
		sep = "/"
	}
	listFile := fmt.Sprintf("%v%v%v", env.Cfg.Seclists, sep, env.Cfg.FuzzDirList)

	w.fuzzFromList(target, env, listFile)
}

func (w *Worker) FuzzFiles() {
	
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
	}
}
