// Joseph Bursey <jbursey@tevora.com>

package httpclient

import (
    "fmt"
    "net/http"
    "time"

    "fafo/pkg/log"
    "fafo/pkg/pretty"
    "fafo/pkg/semaphore"
)

type HttpCfg struct {
    UserAgent string
    MaxCalls  int
    Redirect  func(req *http.Request, via []*http.Request) error
    Timeout   time.Duration
    Slowdown  time.Duration // Time in between consecutive calls (allowing for MaxCalls simultaneous calls)
}

type HttpClient struct {
    client      http.Client
    sem         semaphore.Semaphore
    slowdown    time.Duration
}

func (c HttpClient) Logf(v int, msg string, args ...any) {
    if log.Verb(v) {
        log.Logf(3, "%*s", pretty.PrefixWidth, "[Client]: ")
        log.Logf(v, msg, args...)
    }
}

func (c HttpClient) Log(v int, msg string) {
    c.Logf(v, "%v", msg)
}

func (c HttpClient) Errf(msg string, args ...any) {
    log.Logf(0, fmt.Sprintf("%*s%v: %v", pretty.PrefixWidth, "[Client]: ", pretty.Orange("Error"), msg), args...)
}

func DefaultConfig() *HttpCfg {
    return &HttpCfg{
        UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
        MaxCalls:  5,
        Redirect:  func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
        Timeout:   5*time.Second,
        Slowdown:  200*time.Millisecond,
    }
}

func New(cfg HttpCfg) *HttpClient {
    return &HttpClient{
        client:   http.Client{
            CheckRedirect: cfg.Redirect,
            Timeout:       cfg.Timeout,
        },
        sem:      *semaphore.New(cfg.MaxCalls),
        slowdown: cfg.Slowdown,
    }
}

// Wait for slowdown time until we let another thread in here again
func (c *HttpClient) doSlowdown() {
    time.Sleep(c.slowdown)
    c.sem.Release()
}

func (c *HttpClient) Call(req *http.Request) *http.Response {
    c.sem.Acquire()
    c.Logf(7, "Calling %v\n", req.URL)
    resp, err := c.client.Do(req)
    go c.doSlowdown()
    if err != nil {
        c.Errf("Call to %v failed: %v\n", req.URL, err)
        return nil
    }

    c.Logf(4, "Response: %v\n", resp.Status)
    return resp
}
