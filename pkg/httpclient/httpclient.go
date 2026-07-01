// Joseph Bursey <jbursey@tevora.com>

package httpclient

import (
	"net/http"

	"fafo/pkg/log"
)

type HttpClient struct {
	client http.Client
	// Semaphores and such
}

func (c HttpClient) Log(v int, msg string) {
	c.Logf(v, "%v", msg)
}

func (c HttpClient) Logf(v int, msg string, args ...any) {
	log.Logf(v, "[HttpClient]: "+msg, args...)
}

func (c *HttpClient) Get(url string) *http.Response {
	// Do things with semaphore
	c.Logf(2, "Getting %v\n", url)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		c.Logf(0, "Error in GET request to %v\n", url)
		return nil
	}

	c.Logf(0, "Response: %v\n", resp.Status)
	return resp
}
