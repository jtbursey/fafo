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

    "fafo/pkg/action"
    "fafo/pkg/env"
    "fafo/pkg/fact"
    "fafo/pkg/fs"
    "fafo/pkg/job"
    "fafo/pkg/log"
    "fafo/pkg/pretty"
)

var (
    // Defaults from ffuf
    aliveValid = []int {200, 204, 301, 302, 307, 401}
)

type Fam struct {
    Caller string                // Id of whoever called this (i.e. "Worker 0")
    plch   chan action.Payload
    signal bool
    wg     sync.WaitGroup
}

func (fam *Fam) Logf(v int, msg string, args ...any) {
    prefix := ""
    if log.Verb(3) {
        prefix = fmt.Sprintf("%*s", pretty.PrefixWidth, fmt.Sprintf("[%v]: ", fam.Caller))
    }
    log.Logf(v, prefix+msg, args...)
}

func (fam *Fam) Log(v int, msg string) {
    fam.Logf(v, "%v", msg)
}

func (fam *Fam) Errf(msg string, args ...any) {
    log.Logf(0, fmt.Sprintf("%*s%v: %v", pretty.PrefixWidth, fmt.Sprintf("[%v]: ", fam.Caller), pretty.Orange("Error"), msg), args...)
}

func (fam *Fam) Err(msg string) {
    fam.Errf("%v\n", msg)
}

func (fam *Fam) Init(env *env.Env) {
    fam.signal = false
    fam.plch = make(chan action.Payload, env.Cfg.ClientCfg.MaxCalls*2)
}

func (fam *Fam) channelFile(Pylds *action.PayloadSet, env *env.Env) {
    defer func(){fam.signal = true}()
    filename, err := env.Cfg.GetAsFilename(Pylds.File)
    if err != nil {
        fam.Errf("%v\n", err)
        return
    }
    file, err := os.Open(filename)
    if err != nil {
        fam.Errf("Failed to open file %v: %v\n", filename, err)
        return
    }

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        fam.plch <- action.Payload{
            Id:   Pylds.Id,
            Pl:   scanner.Text(),
        }
    }
    file.Close()
}

func (fam *Fam) channelList(Pylds *action.PayloadSet) {
    defer func(){fam.signal = true}()
    for _, pl := range Pylds.List {
        fam.plch <- action.Payload{
            Id:   Pylds.Id,
            Pl:   pl,
        }
    }
}

func (fam *Fam) channelEmptyPayload() {
    fam.plch <- action.Payload{
        Id:   "",
        Pl:   "",
    } // Default signals work directly on target, no payload
    fam.signal = true
}

func (fam *Fam) channelPayload(pylds *action.PayloadSet, e *env.Env) int {
    var err error
    count := 1
    if pylds == nil { // Handle nil first to prevent issues
        fam.channelEmptyPayload()
    } else if pylds.File != "" {
        count, err = fs.Wc(pylds.File)
        if err != nil {
            fam.Errf("Failed to open file %v: %v\n", pylds.File, err)
            return -1
        }
        fam.wg.Go(func() {fam.channelFile(pylds, e)})
    } else if pylds.List != nil {
        count = len(pylds.List)
        fam.wg.Go(func() {fam.channelList(pylds)})
    } else {
        fam.channelEmptyPayload()
    }
    return count
}

func (fam *Fam) buildMethod(pyld *action.Payload, reqt *action.RequestTemplate) string {
    method := reqt.Method
    if len(pyld.Id) > 0 && strings.Contains(method, pyld.Id) {
        strings.ReplaceAll(method, pyld.Id, pyld.Pl)
    }
    return method
}

func (fam *Fam) buildUrl(pyld *action.Payload, base *fact.Target, reqt *action.RequestTemplate) string {
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

func (fam *Fam) buildBodyReader(pyld *action.Payload, base *fact.Target, reqt *action.RequestTemplate) io.Reader {
    return nil
}

// For now the request is simple. No need for much
func (fam *Fam) buildRequest(pyld *action.Payload, base *fact.Target, reqt *action.RequestTemplate) *http.Request {
    url := fam.buildUrl(pyld, base, reqt)
    req, _ := http.NewRequest(fam.buildMethod(pyld, reqt), url, fam.buildBodyReader(pyld, base, reqt))
    if req == nil {
        fam.Errf("Failed to build request for %v (Base: %v)\n", url)
        return nil
    }
    return req
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

func (fam *Fam) handleResponse(resp *http.Response, req *http.Request, base *fact.Target, respAct *action.ResponseAction, env *env.Env) {
    // Until we actually parse the body...
    _, err := io.ReadAll(resp.Body)
    resp.Body.Close()
    if err != nil {
        fam.Errf("Unexpected error in reading response body: %v\n", err)
    }

    res := fact.Target{
        Url:   resp.Request.URL.String(), // Use the final URL
        Port:  base.Port,
        Facts: make(map[fact.FactKey]fact.FactValue),
    }

    if !env.Cfg.DisableScreenShot && respAct.ScrShcond != nil {
        if respAct.ScrShcond.Evaluate(resp, req, base, &env.Cfg) {
            env.ScrShCh <- *resp.Request
        }
    }

    if slices.Contains(aliveValid, resp.StatusCode) {
        fam.Logf(0, "%v\n", pretty.Response(resp, req.URL.String()))
    } else {
        fam.Logf(2, "%v\n", pretty.Response(resp, res.Url))
    }

    // Push Facts
    for _, pair := range respAct.Factcond {
        if pair.Fingerprint.Evaluate(resp, req, base, &env.Cfg) {
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
        if pair.Fingerprint.Evaluate(resp, req, base, &env.Cfg) {
            for _, j := range pair.Jobs {
                env.JobCh <- fam.buildJob(&j, &res)
            }
        }
    }
}

func (fam *Fam) handlePayload(pyld *action.Payload, base *fact.Target, action *action.Action, env *env.Env) {
    req := fam.buildRequest(pyld, base, action.Reqt)
    if req == nil {
        return
    }

    // TODO: Figure out logic to tell the fuzzer to not Call
    resp := env.Client.Call(req)
    if resp == nil {
        fam.Err("Call Failed!")
        return
    }

    fam.handleResponse(resp, req, base, action.RespAct, env)
}

func (fam *Fam) childLoop(b *fact.Target, a *action.Action, e *env.Env) {
    for {
        select {
        case pyld := <- fam.plch:
            fam.handlePayload(&pyld, b, a, e)
        default:
            if fam.signal {return}
        }
    }
}

func (fam *Fam) runChildren(b *fact.Target, a *action.Action, env *env.Env, count int) {
    // Take hint from max calls for how many children to spawn
    for range min(env.Cfg.ClientCfg.MaxCalls, count) {
        fam.wg.Go(func() {fam.childLoop(b, a, env)})
    }
}

func (fam *Fam) Run(b *fact.Target, action *action.Action, e *env.Env) {
    fam.Init(e)

    // Handle Payload
    count := fam.channelPayload(action.Pylds, e)
    if count < 0 {
        return
    }

    // Spawn Children
    fam.runChildren(b, action, e, count)

    // Wait for it all to finish
    fam.wg.Wait()
}
