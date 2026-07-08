// Joseph Bursey <jbursey@tevora.com>

// Fuzz Anything Machine

package fam

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"fafo/pkg/env"
	"fafo/pkg/fact"
	"fafo/pkg/job"
	"fafo/pkg/log"
	"fafo/pkg/pretty"
)

var (
	// Defaults from ffuf
	aliveValid = []int {200, 204, 301, 302, 307, 401}
)

type Fam struct {
	Caller string				// Id of whoever called this (i.e. "Worker 0")
	plch   chan Payload
	signal bool
	wg     sync.WaitGroup
}

type FamRequest struct {
	Req     *http.Request
	Type    fact.TargetType
}

// write something in the profile/config/whatever
	// that file can be id'ed by filename or internal id after parsed to memory
// Some previous action knows what next action should be taken
// It passes th relevant filename/id to FAM
// FAM reads the object/file and takes actions based on it.

// Generalize
	// Base target
	// Channel Payloads

	// In Minion:
		// Build http request
			// Inject the Payload (in URL, Body, Header, etc.)
			// Custom request options
		// Send request
		// Receive response
		// Do something with body (or drop)
			// Custom matching eventually
		// Return Action
			// Push what job based on what observation
			// Push what Fact based on what observation

func (fam *Fam) Logf(v int, msg string, args ...any) {
	prefix := ""
	if log.Verb(3) {
		prefix = fmt.Sprintf("%-13v", fmt.Sprintf("[%v]: ", fam.Caller))
	}
	log.Logf(v, prefix+msg, args...)
}

func (fam *Fam) Log(v int, msg string) {
	fam.Logf(v, "%v", msg)
}

func (fam *Fam) Errf(msg string, args ...any) {
	log.Logf(0, fmt.Sprintf("%-13v%v: %v\n", fmt.Sprintf("[%v]: ", fam.Caller), pretty.Orange("Error"), msg), args...)
}

func (fam *Fam) Err(msg string) {
	fam.Errf("%v", msg)
}

func (fam *Fam) Init(env *env.Env) {
	fam.signal = false
	fam.plch = make(chan Payload, env.Cfg.ClientCfg.MaxCalls*2)
}

func wc(filename string) int {
	count := 0
	file, err := os.Open(filename)
	if err != nil {
		log.Errf("Failed to open file %v: %v.", filename, err)
		return -1
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	file.Close()
	return count
}

func (fam *Fam) channelFile(Pylds *PayloadSet) {
	file, err := os.Open(Pylds.File)
	if err != nil {
		fam.Errf("Failed to open file %v: %v.", Pylds.File, err)
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fam.plch <- Payload{
			Id:   Pylds.Id,
			Type: Pylds.Type,
			Pl:   scanner.Text(),
		}
	}
	file.Close()
	fam.signal = true
}

func (fam *Fam) channelList(Pylds *PayloadSet) {
	for _, pl := range Pylds.List {
		fam.plch <- pl
	}
	fam.signal = true
}

func (fam *Fam) channelPayload(pylds *PayloadSet) int {
	count := 0
	if pylds.File != "" {
		count = wc(pylds.File)
		fam.wg.Go(func() {fam.channelFile(pylds)})
	} else if pylds.List != nil {
		count = len(pylds.List)
		fam.wg.Go(func() {fam.channelList(pylds)})
	} else {
		count = 1
		fam.plch <- Payload{
			Id:   "",
			Type: pylds.Type,
			Pl:   "",
		} // Default signals work directly on target
		fam.signal = true
	}
	return count
}

func (fam *Fam) buildMethod(pyld *Payload, reqt *RequestTemplate) string {
	method := reqt.Method
	if len(pyld.Id) > 0 && strings.Contains(method, pyld.Id) {
		strings.ReplaceAll(method, pyld.Id, pyld.Pl)
	}
	return method
}

func (fam *Fam) buildUrl(pyld *Payload, base *fact.Target, reqt *RequestTemplate) string {
	url := reqt.Url
	if strings.Contains(url, "BASE") {
		url = strings.ReplaceAll(url, "BASE", base.Url)
	} else {
		fam.Err("No BASE in Url Template!")
		return ""
	}

	if len(pyld.Id) > 0 && strings.Contains(url, pyld.Id) {
		url = strings.ReplaceAll(url, pyld.Id, pyld.Pl)
	}

	return url
}

func (fam *Fam) buildBodyReader(pyld *Payload, base *fact.Target, reqt *RequestTemplate) io.Reader {
	return nil
}

// For now the request is simple. No need for much
func (fam *Fam) buildRequest(pyld *Payload, base *fact.Target, reqt *RequestTemplate) *FamRequest {
	req, _ := http.NewRequest(fam.buildMethod(pyld, reqt), fam.buildUrl(pyld, base, reqt), fam.buildBodyReader(pyld, base, reqt))
	if req == nil {
		fam.Err("Failed to build request!")
		return nil
	}
	return &FamRequest{
		Req:  req,
		Type: pyld.Type,
	}
}

func (fam *Fam) buildJob(base *job.Job, target *fact.Target) job.Job {
	newJob := job.Job{
		Mode:     base.Mode,
		Action:   base.Action,
		Priority: base.Priority,
	}

	if base.Target == "CURRENT" {
		newJob.Target = target.Key()
	} else {
		fam.Err("Unspecified Target for new job")
	}

	return newJob
}

func (fam *Fam) handleResponse(resp *http.Response, req *FamRequest, base *fact.Target, respAct *ResponseAction, env *env.Env) {
	// Until we actually parse the body...
	_, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fam.Errf("Unexpected error in reading response body: %v", err)
	}

	res := fact.Target{
		Url:   req.Req.URL.String(),
		Type:  req.Type,
		Port:  base.Port,
		Facts: make(map[fact.FactKey]fact.FactValue),
	}

	if slices.Contains(aliveValid, resp.StatusCode) {
		fam.Logf(0, "%v\n", pretty.Response(resp, req.Req.URL.String()))
	} else {
		fam.Logf(2, "%v\n", pretty.Response(resp, res.Url))
	}

	// Push Facts
	for _, pair := range respAct.Factcond {
		if pair.Fingerprint.Evaluate(resp, req, base) {
			for key, value := range pair.FactPair {
				res.Facts[key] = value
			}
		}
	}

	if len(res.Facts) > 0 {
		env.FactCh <- res
	}

	// Push Jobs
	for _, pair := range respAct.Jobcond {
		if pair.Fingerprint.Evaluate(resp, req, base) {
			env.JobCh <- fam.buildJob(&pair.Job, &res)
		}
	}
}

func (fam *Fam) handlePayload(pyld *Payload, base *fact.Target, action *Action, env *env.Env) {
	req := fam.buildRequest(pyld, base, action.Reqt)

	// TODO: Figure out logic to tell the fuzzer to not Call
	resp := env.Client.Call(req.Req)
	if resp == nil {
		fam.Err("Call Failed!")
		return
	}

	fam.handleResponse(resp, req, base, action.RespAct, env)
}

func (fam *Fam) childLoop(b *fact.Target, a *Action, e *env.Env) {
	for {
		select {
		case pyld := <- fam.plch:
			fam.handlePayload(&pyld, b, a, e)
		default:
			if fam.signal {return}
		}
	}
}

func (fam *Fam) runChildren(b *fact.Target, a *Action, env *env.Env, count int) {
	// Take hint from max calls for how many children to spawn
	for range min(env.Cfg.ClientCfg.MaxCalls, count) {
		fam.wg.Go(func() {fam.childLoop(b, a, env)})
	}
}

func (fam *Fam) Run(b *fact.Target, action *Action, e *env.Env) {
	fam.Init(e)

	// Handle Payload
	count := fam.channelPayload(action.Pylds)
	if count < 0 {
		return
	}

	// Spawn Children
	fam.runChildren(b, action, e, count)

	// Wait for it all to finish
	fam.wg.Wait()
}
