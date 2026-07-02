// Joseph Bursey <jbursey@tevora.com>

package httpclient

import (
	"net/http"

	"fafo/pkg/log"
	"fafo/pkg/pretty"
)

type HttpClient struct {
	client http.Client
	// Semaphores and such
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

func (c *HttpClient) Get(url string) *http.Response {
	// Do things with semaphore
	c.Logf(4, "Getting %v\n", url)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		c.Errf("GET request to %v failed.\n", url)
		return nil
	}

	c.Logf(4, "Response: %v\n", resp.Status)
	return resp
}
