// Joseph Bursey <jbursey@tevora.com>

package httpclient

import (
    "crypto/tls"
    "fmt"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "time"

    "fafo/pkg/log"
    "fafo/pkg/pretty"
    "fafo/pkg/semaphore"
)

type HttpCfg struct {
    UserAgent  string        `json:"UserAgent"`
    // TODO: Figure out a way to unmarshall json to url
    Proxy      *url.URL
    MaxCalls   int           `json:"MaxCalls"`
    doRedirect bool          `json:"FollowRedirects"`
    Redirect   func(req *http.Request, via []*http.Request) error
    Timeout    time.Duration `json:"Timeout"`
    Slowdown   time.Duration `json:"Slowdown"`
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
        UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
        MaxCalls:   5,
        doRedirect: false,
        Redirect:   func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
        Timeout:    5000*time.Millisecond,
        Slowdown:   200*time.Millisecond,
    }
}

func (c *HttpCfg) PostParse() {
    if !c.doRedirect {
        c.Redirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
    }

    c.Slowdown = c.Slowdown * time.Millisecond
    c.Timeout = c.Timeout * time.Millisecond
}

func (c *HttpCfg) Debug() {
    if c.Proxy != nil {
        log.Logf(0, "%v\n", pretty.Config("Proxy", c.Proxy.String()))
    }
    log.Logf(0, "%v\n", pretty.Config("MaxCalls", c.MaxCalls))
    log.Logf(0, "%v\n", pretty.Config("Slowdown", c.Slowdown))
    log.Logf(0, "%v\n", pretty.Config("Timeout", c.Timeout))
    log.Logf(0, "%v\n", pretty.Config("FollowRedirects", c.doRedirect))
    log.Logf(0, "%v\n", pretty.Config("UserAgent", c.UserAgent))
}

func New(cfg HttpCfg) *HttpClient {
    jar, err := cookiejar.New(nil)
    if err != nil {
        log.Errf("Failed to cookie jar: %v", err)
        return nil
    }
    client := &HttpClient{
        client:   http.Client{
            CheckRedirect: cfg.Redirect,
            Timeout:       cfg.Timeout,
            Jar:           jar,
        },
        sem:      *semaphore.New(cfg.MaxCalls),
        slowdown: cfg.Slowdown,
    }

    if cfg.Proxy != nil {
        client.client.Transport = &http.Transport{
            Proxy: http.ProxyURL(cfg.Proxy),
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        }
    }

    return client
}

// Wait for slowdown time until we let another thread in here again
func (c *HttpClient) doSlowdown() {
    time.Sleep(c.slowdown)
    c.sem.Release()
}

func (c *HttpClient) BorrowSem() {
    c.sem.Acquire()
}

func (c *HttpClient) ReturnSem() {
    go c.doSlowdown()
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
