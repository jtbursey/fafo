// Joseph Bursey <jbursey@tevora.com>

// Fuzz Anything Machine

package fam

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
    "net/url"
	"os"
	"slices"
	"strings"
	"sync"

	"fafo/pkg/action"
	"fafo/pkg/env"
	"fafo/pkg/fact"
    "fafo/pkg/fingerprint"
	"fafo/pkg/fs"
	"fafo/pkg/httpclient"
	"fafo/pkg/job"
	"fafo/pkg/log"
	"fafo/pkg/pretty"
)

var (
    // Defaults from ffuf
    aliveValid = []int {200, 204, 301, 302, 307, 401, 405}
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

func (fam *Fam) Warnf(msg string, args ...any) {
    log.Logf(2, fmt.Sprintf("%*s%v: %v", pretty.PrefixWidth, fmt.Sprintf("[%v]: ", fam.Caller), pretty.Orange("Warning"), msg), args...)
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
    count := 1
    if pylds == nil { // Handle nil first to prevent issues
        fam.channelEmptyPayload()
    } else if pylds.File != "" {
        file, err := e.Cfg.GetAsFilename(pylds.File)
        if err != nil {
            fam.Errf("%v\n", err)
            return -1
        }
        count, err = fs.Wc(file)
        if err != nil {
            fam.Errf("Failed to open file %v: %v\n", file, err)
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

func (fam *Fam) payloadReplace(pyld *action.Payload, origin string) string {
    if len(pyld.Id) > 0 && strings.Contains(origin, pyld.Id) {
        origin = strings.ReplaceAll(origin, pyld.Id, pyld.Pl)
    }
    return origin
}

func (fam *Fam) buildMethod(pyld *action.Payload, reqt *action.RequestTemplate) string {
    return fam.payloadReplace(pyld, reqt.Method)
}

func (fam *Fam) buildUrl(pyld *action.Payload, base *fact.Target, reqt *action.RequestTemplate) (*url.URL, error) {
    newUrl := reqt.Url
    if strings.Contains(newUrl, "BASE") {
        newUrl = strings.ReplaceAll(newUrl, "BASE", base.Url.String())
    } else {
        return nil, fmt.Errorf("No BASE in Url Template: %v", newUrl)
    }

    ret, err := url.Parse(fam.payloadReplace(pyld, newUrl))
    if err != nil {
        return nil, fmt.Errorf("Failed to parse new Url: %v: %v\n", fam.payloadReplace(pyld, newUrl), err)
    }
    return ret, nil
}

func (fam *Fam) buildBodyReader(pyld *action.Payload, base *fact.Target, reqt *action.RequestTemplate) io.Reader {
    if reqt.Body == nil {
        return nil
    }
    body := strings.Join(reqt.Body, "\r\n")
    body = fam.payloadReplace(pyld, body)
    return strings.NewReader(body)
}

func (fam *Fam) buildHeader(pyld *action.Payload, reqt *action.RequestTemplate, cfg *httpclient.HttpCfg) map[string][]string {
    header := make(map[string][]string)
    for hdr, val := range reqt.Header {
        if hdr == "User-Agent" && val == "DEFAULT" {
            header["User-Agent"] = []string{cfg.UserAgent}
            continue
        }
        header[hdr] = append(header[hdr], fam.payloadReplace(pyld, val))
    }

    if reqt.Header == nil || reqt.Header["User-Agent"] == "" {
        header["User-Agent"] = []string{cfg.UserAgent}
    }

    return header
}

// For now the request is simple. No need for much
func (fam *Fam) buildRequest(pyld *action.Payload, base *fact.Target, reqt *action.RequestTemplate, env *env.Env) *http.Request {
    url, err := fam.buildUrl(pyld, base, reqt)
    if err != nil {
        fam.Errf("%v\n", err)
        return nil
    }
    req, _ := http.NewRequest(fam.buildMethod(pyld, reqt), url.String(), fam.buildBodyReader(pyld, base, reqt))
    if req == nil {
        fam.Errf("Failed to build request for %v (Base: %v)\n", url.String(), base.Url.String())
        return nil
    }

    req.Header = fam.buildHeader(pyld, reqt, &env.Cfg.ClientCfg)
    return req
}

func (fam *Fam) buildJob(base *job.Job, target *fact.Target) job.Job {
    newJob := job.Job{
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
        fam.Warnf("Unexpected error in reading response body: %v\n", err)
    }

    res := fact.Target{
        Url:   resp.Request.URL, // Use the final URL
        Facts: make(map[fact.FactKey]fact.FactValue),
    }

    if !env.Cfg.DisableScreenShot && respAct.ScrShcond != nil {
        b, err := respAct.ScrShcond.Evaluate(resp, req, base, &env.Cfg)
        if err != nil {
            fam.Warnf("Failed to evaluation Screenshot condition: %v", err)
        }
        if b {
            env.ScrShCh <- *resp.Request
        }
    }

    if slices.Contains(aliveValid, resp.StatusCode) {
        fam.Logf(0, "%v\n", pretty.Response(resp, req.URL.String()))
    } else {
        fam.Logf(2, "%v\n", pretty.Response(resp, res.Url.String()))
    }

    // Push Facts
    for _, pair := range respAct.Factcond {
        b, err := pair.Fingerprint.Evaluate(resp, req, base, &env.Cfg)
        if err != nil {
            fam.Errf("Failed to evaluation Fact condition: %v\n", err)
        }
        if b {
            for key, value := range pair.FactPair {
                if val, err := fingerprint.Field(value).Get(resp, req, base, &env.Cfg); err == nil {
                    res.Facts[key] = fact.FactValue(val)
                    continue
                } else {
                    res.Facts[key] = value
                }
            }
        }
    }

    if len(res.Facts) > 0 {
        env.FactCh <- res
    }

    // Push Jobs
    for _, pair := range respAct.Jobcond {
        b, err := pair.Fingerprint.Evaluate(resp, req, base, &env.Cfg)
        if err != nil {
            fam.Errf("Failed to evaluation Job condition: %v\n", err)
        }
        if b {
            for _, j := range pair.Jobs {
                env.JobCh <- fam.buildJob(&j, &res)
            }
        }
    }
}

func (fam *Fam) handlePayload(pyld *action.Payload, base *fact.Target, action *action.Action, env *env.Env) {
    req := fam.buildRequest(pyld, base, action.Reqt, env)
    if req == nil {
        return
    }

    // TODO: try to pull the new request as a target to see if we've called it already.
        // Need to figure out what kinds of conditionals should considered

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
            // This stops one hell of a race.
                // 1. child spawns and sees there is no payload to pull
                // 2. channeler channels the payload and signals done
                // 3. child goes to default and checks if done, returns.
            if fam.signal && len(fam.plch) == 0 {return}
        }
    }
}

func (fam *Fam) runChildren(b *fact.Target, a *action.Action, env *env.Env, count int) {
    // Take hint from max calls for how many children to spawn
    childCount := min(env.Cfg.ClientCfg.MaxCalls, count)
    for range childCount {
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
