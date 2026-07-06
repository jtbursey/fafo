// Joseph Bursey <jbursey@tevora.com>

package httpclient

import (
	"io"
	"net/http"
	"time"

	"fafo/pkg/log"
	"fafo/pkg/pretty"
	"fafo/pkg/semaphore"
)

type HttpCfg struct {
	MaxCalls int
	Redirect func(req *http.Request, via []*http.Request) error
	Timeout  time.Duration
}

type HttpClient struct {
	client http.Client
	sem    semaphore.Semaphore
}

func (c HttpClient) Logf(v int, msg string, args ...any) {
	if log.Verb(v) {
		log.Logf(3, "%-13v", "[Client]: ")
		log.Logf(v, msg, args...)
	}
}

func (c HttpClient) Log(v int, msg string) {
	c.Logf(v, "%v", msg)
}

func (c HttpClient) Errf(msg string, args ...any) {
	log.Logf(0, "%-13v %v: %v", append([]any{"[Client]:", pretty.Orange("Error"), msg}, args...))
}

func DefaultConfig() *HttpCfg {
	return &HttpCfg{
		MaxCalls:  5,
		Redirect:  func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:   time.Second,
	}
}

func New(cfg HttpCfg) *HttpClient {

	return &HttpClient{
		client: http.Client{
			CheckRedirect: cfg.Redirect,
			Timeout:       cfg.Timeout,
		},
		sem:    *semaphore.New(cfg.MaxCalls),
	}
}

func (c *HttpClient) Get(url string) *http.Response {
	c.sem.Acquire()
	c.Logf(7, "Getting %v\n", url)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := c.client.Do(req)
	c.sem.Release()
	if err != nil {
		c.Errf("GET request to %v failed.\n", url)
		return nil
	}

	c.Logf(4, "Response: %v\n", resp.Status)
	return resp
}

// Just read the body and close it out. Call if you don't care about the body.
func (c *HttpClient) DropBody(resp *http.Response) {
	_, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		c.Errf("Unexpected error in DropBody: %v\n", err)
	}
}
