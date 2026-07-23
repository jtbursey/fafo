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
    plch   chan []action.Payload
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
    fam.plch = make(chan []action.Payload, env.Cfg.ClientCfg.MaxCalls*2)
}

func (fam *Fam) countPayloads(pylds []action.PayloadOrigin, e *env.Env) (int, error) {
    count := 1
    if pylds == nil {
        return count, nil
    }
    for _, origin := range pylds {
        c := 1
        if origin.File != "" {
            file, err := e.Cfg.GetAsFilename(origin.File)
            if err != nil {
                fam.Errf("%v\n", err)
                return count, err
            }
            c, err = fs.Wc(file)
            if err != nil {
                fam.Errf("Failed to open file %v: %v\n", file, err)
                return count, err
            }
        } else if origin.List != nil {
            c = len(origin.List)
        }
        // else count = 1
        count *= c
    }

    return count, nil
}

func (fam *Fam) recursiveChannel(list []action.PayloadOrigin, curPylds []action.Payload, env *env.Env) error {
    // TODO: turn this into sub-functions

    if len(list) == 0 && len(curPylds) == 0 {
        newPylds := append(curPylds, action.Payload{
            Id:   "",
            Pl:   "",
        })
        fam.plch <- newPylds
        return nil
    } else if len(list) == 0 {
        fam.Warnf("Called recursiveChannel on empty list.\n")
        return nil
    }

    current := list[0]
    list = list[1:]
    if current.File != "" {
        filename, err := env.Cfg.GetAsFilename(current.File)
        if err != nil {
            return err
        }
        file, err := os.Open(filename)
        if err != nil {
            return err
        }

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            newPylds := append(curPylds, action.Payload{
                Id:   current.Id,
                Pl:   scanner.Text(),
            })
            if len(list) > 0 {
                if err := fam.recursiveChannel(list, newPylds, env); err != nil {
                    return err
                }
            } else {
                fam.plch <- newPylds
            }
        }
        file.Close()
    } else if current.List != nil {
        for _, pl := range current.List {
            newPylds := append(curPylds, action.Payload{
                Id:   current.Id,
                Pl:   pl,
            })
            if len(list) > 0 {
                if err := fam.recursiveChannel(list, newPylds, env); err != nil {
                    return err
                }
            } else {
                fam.plch <- newPylds
            }
        }
    } else {
        newPylds := append(curPylds, action.Payload{
            Id:   current.Id,
            Pl:   "",
        })
        if len(list) > 0 {
            if err := fam.recursiveChannel(list, newPylds, env); err != nil {
                return err
            }
        } else {
            fam.plch <- newPylds
        }
    }

    return nil
}

func (fam *Fam) channelPayloads(pylds []action.PayloadOrigin, e *env.Env) (int, error) {
    count, err := fam.countPayloads(pylds, e)
    if err != nil {
        fam.Errf("Failed to count Payloads: %v", err)
        return count, err
    }

    fam.wg.Go(func() {fam.recursiveChannel(pylds, make([]action.Payload, 0), e); fam.signal = true})

    return count, nil
}

func (fam *Fam) payloadReplace(pyldList []action.Payload, origin string) string {
    for _, pyld := range pyldList {
        if len(pyld.Id) > 0 && strings.Contains(origin, pyld.Id) {
            origin = strings.ReplaceAll(origin, pyld.Id, pyld.Pl)
        }
    }
    
    return origin
}

func (fam *Fam) baseReplace(base *fact.Target, origin string) (string, error) {
    if strings.Contains(origin, "BASE") {
        origin = strings.ReplaceAll(origin, "BASE", base.Url.String())
    } else {
        return origin, fmt.Errorf("No BASE in Url Template: %v", origin)
    }
    return origin, nil
}

func (fam *Fam) fullReplace(pyldList []action.Payload, base *fact.Target, origin string) string {
    origin, _ = fam.baseReplace(base, origin)

    if strings.Contains(origin, "CURRENT") {
        origin = strings.ReplaceAll(origin, "CURRENT", base.Url.String())
    }

    origin = fam.payloadReplace(pyldList, origin)
    return origin
}

func (fam *Fam) buildMethod(pyld []action.Payload, reqt *action.RequestTemplate) string {
    return fam.payloadReplace(pyld, reqt.Method)
}

func (fam *Fam) buildUrl(pyld []action.Payload, base *fact.Target, reqt *action.RequestTemplate) (*url.URL, error) {
    newUrl := reqt.Url
    var err error
    if newUrl, err = fam.baseReplace(base, newUrl); err != nil {
        return nil, err
    }

    ret, err := url.Parse(fam.payloadReplace(pyld, newUrl))
    if err != nil {
        return nil, fmt.Errorf("Failed to parse new Url: %v: %v\n", fam.payloadReplace(pyld, newUrl), err)
    }
    return ret, nil
}

func (fam *Fam) buildBodyReader(pyld []action.Payload, base *fact.Target, reqt *action.RequestTemplate) io.Reader {
    if reqt.Body == nil {
        return nil
    }
    body := strings.Join(reqt.Body, "\r\n")
    body = fam.payloadReplace(pyld, body)
    body += "\r\n\r\n"
    return strings.NewReader(body)
}

func (fam *Fam) buildHeader(pyld []action.Payload, reqt *action.RequestTemplate, cfg *httpclient.HttpCfg) map[string][]string {
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
func (fam *Fam) buildRequest(pyld []action.Payload, base *fact.Target, reqt *action.RequestTemplate, env *env.Env) *http.Request {
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

func (fam *Fam) buildJob(pyld []action.Payload, base *job.Job, target *fact.Target) job.Job {
    newJob := job.Job{
        Action:   base.Action,
        Priority: base.Priority,
    }

    newJob.Target = fam.fullReplace(pyld, target, base.Target)
    if newJob.Target == "" {
        fam.Err("Unspecified Target for new job")
    }

    return newJob
}

func (fam *Fam) handleResponse(pyld []action.Payload, resp *http.Response, req *http.Request, base *fact.Target, respAct *action.ResponseAction, env *env.Env) {
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

    // TODO: print the payloads here
    if slices.Contains(aliveValid, resp.StatusCode) {
        fam.Logf(0, "%v\n", pretty.Response(resp, req.URL.String()))
    } else if resp.StatusCode != 404 {
        fam.Logf(1, "%v\n", pretty.Response(resp, req.URL.String()))
    } else {
        fam.Logf(2, "%v\n", pretty.Response(resp, res.Url.String()))
    }

    // Push Facts
    for _, pair := range respAct.Factcond {
        b, err := pair.Fingerprint.Evaluate(resp, req, base, &env.Cfg)
        if err != nil {
            fam.Warnf("Failed to evaluation Fact condition: %v\n", err)
        }
        if b {
            for key, value := range pair.FactPair {
                if val, err := fingerprint.Field(value).Get(resp, req, base, &env.Cfg); err == nil {
                    res.Facts[key] = fact.FactValue(val)
                    continue
                } else {
                    res.Facts[key] = fact.FactValue(fam.payloadReplace(pyld, string(value)))
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
            fam.Warnf("Failed to evaluation Job condition: %v\n", err)
        }
        if b {
            for _, j := range pair.Jobs {
                env.JobCh <- fam.buildJob(pyld, &j, &res)
            }
        }
    }
}

func (fam *Fam) handlePayload(pyld []action.Payload, base *fact.Target, action *action.Action, env *env.Env) {
    req := fam.buildRequest(pyld, base, action.Reqt, env)
    if req == nil {
        return
    }

    // TODO: Maybe allow for multiple calls to set cookies?
        // TODO: This also means multiple http clients for different cookies

    // TODO: try to pull the new request as a target to see if we've called it already.
        // Need to figure out what kinds of conditionals should considered

    // TODO: Figure out logic to tell the fuzzer to not Call
    resp := env.Client.Call(req)
    if resp == nil {
        fam.Err("Call Failed!")
        return
    }

    fam.handleResponse(pyld, resp, req, base, action.RespAct, env)
}

func (fam *Fam) childLoop(b *fact.Target, a *action.Action, e *env.Env) {
    for {
        select {
        case pyld := <- fam.plch:
            fam.handlePayload(pyld, b, a, e)
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
    count, err := fam.channelPayloads(action.Pylds, e)
    if err != nil {
        fam.Errf("Failed while channeling payloads: %v", err)
        return
    }

    // Spawn Children
    fam.runChildren(b, action, e, count)

    // Wait for it all to finish
    fam.wg.Wait()
}
