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

const (
    VError int = log.VError
    VWarn  int = log.VWarn

    VPos   int = log.V0
    V404   int = log.V1
    VOther int = log.V2
)

var (
    // Defaults from ffuf
    aliveValid = []int {200, 204, 301, 302, 307, 401, 405}
)

type Fam struct {
    Caller     string                // Id of whoever called this (i.e. "Worker 0")
    plch       chan []action.Payload
    signal     bool
    wg         sync.WaitGroup
}

func (fam *Fam) prefix() string {
    if log.Verb(log.VPrefix) {
        return fmt.Sprintf("%*s", pretty.PrefixWidth, fmt.Sprintf("[%v]: ", fam.Caller))
    }
    return ""
}

func (fam *Fam) Logf(v int, msg string, args ...any) {
    log.Logf(v, fam.prefix()+msg, args...)
}

func (fam *Fam) Log(v int, msg string) {
    fam.Logf(v, "%v", msg)
}

func (fam *Fam) Errf(msg string, args ...any) {
    log.Logf(VError, fmt.Sprintf("%v%v: %v", fam.prefix(), pretty.Orange("Error"), msg), args...)
}

func (fam *Fam) Err(msg string) {
    fam.Errf("%v\n", msg)
}

func (fam *Fam) Warnf(msg string, args ...any) {
    log.Logf(VWarn, fmt.Sprintf("%v%v: %v", fam.prefix(), pretty.Orange("Warning"), msg), args...)
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

func (fam *Fam) buildMethod(reqt *action.RequestTemplate, ds *Data) string {
    return ds.Replace(reqt.Method)
}

func (fam *Fam) buildUrl(reqt *action.RequestTemplate, ds *Data) (*url.URL, error) {
    newUrl := ds.Replace(reqt.Url)
    ret, err := url.Parse(newUrl)
    if err != nil {
        return nil, fmt.Errorf("Failed to parse new Url: %v: %v\n", newUrl, err)
    }
    return ret, nil
}

func (fam *Fam) buildBodyReader(reqt *action.RequestTemplate, ds *Data) io.Reader {
    if reqt.Body == nil {
        return nil
    }
    body := strings.Join(reqt.Body, "\r\n")
    body = ds.Replace(body)
    body += "\r\n\r\n"
    return strings.NewReader(body)
}

func (fam *Fam) buildHeader(reqt *action.RequestTemplate, ds *Data) map[string][]string {
    header := make(map[string][]string)
    //header["Connection"] = []string{"close"}
    for hdr, val := range reqt.Header {
        if hdr == "User-Agent" && val == "DEFAULT" {
            header["User-Agent"] = []string{ds.Config.ClientCfg.UserAgent}
            continue
        }
        header[hdr] = append(header[hdr], ds.Replace(val))
    }

    if reqt.Header == nil || reqt.Header["User-Agent"] == "" {
        header["User-Agent"] = []string{ds.Config.ClientCfg.UserAgent}
    }

    return header
}

// For now the request is simple. No need for much
func (fam *Fam) buildRequest(reqt *action.RequestTemplate, ds *Data) *http.Request {
    url, err := fam.buildUrl(reqt, ds)
    if err != nil {
        fam.Errf("%v\n", err)
        return nil
    }
    req, _ := http.NewRequest(fam.buildMethod(reqt, ds), url.String(), fam.buildBodyReader(reqt, ds))
    if req == nil {
        fam.Errf("Failed to build request for %v\n", url.String())
        return nil
    }

    req.Header = fam.buildHeader(reqt, ds)
    return req
}

func (fam *Fam) buildJob(baseJob *job.Job, ds *Data) job.Job {
    newJob := job.Job{
        Action:   baseJob.Action,
        Priority: baseJob.Priority,
    }

    newJob.Target = ds.Replace(baseJob.Target)
    if newJob.Target == "" {
        fam.Err("Unspecified Target for new job")
    }

    return newJob
}

func (fam *Fam) handleResponse(respAct *action.ResponseAction, ds *Data, env *env.Env) {
    // Until we actually parse the body...
    _, err := io.ReadAll(ds.Response.Body)
    ds.Response.Body.Close()
    if err != nil {
        fam.Warnf("Unexpected error in reading response body: %v\n", err)
    }

    res := fact.Target{
        Url:   ds.Response.Request.URL, // Use the final URL
        Facts: make(map[fact.FactKey][]string),
    }

    if !ds.Config.DisableScreenShot && respAct.ScrShcond != nil {
        b, err := respAct.ScrShcond.Evaluate(ds.Response, ds.Request, ds.Config)
        if err != nil {
            fam.Warnf("Failed to evaluation Screenshot condition: %v", err)
        }
        if b {
            env.ScrShCh <- *ds.Response.Request
        }
    }

    // Push Facts
    for _, pair := range respAct.Factcond {
        b, err := pair.Fingerprint.Evaluate(ds.Response, ds.Request, ds.Config)
        if err != nil {
            fam.Warnf("Failed to evaluation Fact condition: %v\n", err)
        }
        if b {
            for key, value := range pair.FactPair {
                res.AppendUniqueValues(key, ds.PrepareAppend(string(value)))
            }
        }
    }

    // TODO: print the payloads here, and the Action
    if len(res.Facts) > 0 {
        fam.Logf(VPos, "%v\n", pretty.Response(ds.Response, ds.Request.URL.String()))
        env.FactCh <- res
    } else if ds.Response.StatusCode != 404 {
        fam.Logf(V404, "%v\n", pretty.Response(ds.Response, ds.Request.URL.String()))
    } else {
        fam.Logf(VOther, "%v\n", pretty.Response(ds.Response, ds.Request.URL.String()))
    }

    // Push Jobs
    for _, pair := range respAct.Jobcond {
        b, err := pair.Fingerprint.Evaluate(ds.Response, ds.Request, ds.Config)
        if err != nil {
            fam.Warnf("Failed to evaluation Job condition: %v\n", err)
        }
        if b {
            for _, j := range pair.Jobs {
                env.JobCh <- fam.buildJob(&j, ds)
            }
        }
    }

    // TODO: Allow fam to stop early if attack goal is reached
}

func (fam *Fam) handlePayload(pyld []action.Payload, base *fact.Target, action *action.Action, env *env.Env) {
    ds := NewData()
    ds.TakePayloads(pyld)
    ds.TakeConfig(&env.Cfg)
    ds.TakeBaseUrl(base.Url)
    ds.TakeRequest(fam.buildRequest(action.Reqt, ds))
    if ds.Request == nil {
        return
    }

    // TODO: Maybe allow for multiple calls to set cookies?
        // TODO: This also means multiple http clients for different cookies

    // TODO: try to pull the new request as a target to see if we've called it already.
        // Need to figure out what kinds of conditionals should considered

    // TODO: Figure out logic to tell the fuzzer to not Call
    ds.TakeResponse(env.Client.Call(ds.Request))
    if ds.Response == nil {
        fam.Err("Call Failed!")
        return
    }

    fam.handleResponse(action.RespAct, ds, env)
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
