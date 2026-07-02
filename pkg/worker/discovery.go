// Joseph Bursey <jbursey@tevora.com>

// This file defines various discovery actions that can be carried out by workers

package worker

import (
	"fafo/pkg/env"
	"fafo/pkg/fact"
	//"fafo/pkg/httpclient"
)

// Check alive
func (w *Worker) CheckAlive(target *fact.Target, env env.Env) {
	resp := env.Client.Get(target.Url)
	if resp == nil {
		w.Logf(0, "Error getting target: %v\n", target.Url)
	}

	w.Logf(0, "Received valid response (%v) from %v\n ", resp.StatusCode, target.Url)
	target.IsAlive = 1
}

// Path fuzzing

// WP Scan
