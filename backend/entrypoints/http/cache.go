package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/boreq/rest"
)

func Cache(duration time.Duration, handler rest.HandlerFunc) rest.HandlerFunc {
	return newCache(duration, handler).Handler
}

type cache struct {
	duration time.Duration
	handler  rest.HandlerFunc
	response rest.RestResponse
	recorded time.Time
	mutex    sync.Mutex
}

func newCache(duration time.Duration, handler rest.HandlerFunc) *cache {
	return &cache{
		duration: duration,
		handler:  handler,
	}
}

func (c *cache) Handler(r *http.Request) rest.RestResponse {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.recorded.IsZero() || c.recorded.Add(c.duration).Before(time.Now()) {
		c.recorded = time.Now()
		c.response = c.handler(r)
	}

	return c.response
}
