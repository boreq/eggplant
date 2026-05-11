package http_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	httpPort "github.com/boreq/eggplant/ports/http"
	"github.com/boreq/rest"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	const d = 10 * time.Millisecond
	c := httpPort.Cache(d, makeHandler())
	r1 := c(nil)
	<-time.After(d / 2)
	r2 := c(nil)
	require.Equal(t, r1, r2)
}

func TestCacheInvalidated(t *testing.T) {
	const d = 10 * time.Millisecond
	c := httpPort.Cache(d, makeHandler())
	r1 := c(nil)
	<-time.After(d * 2)
	r2 := c(nil)
	require.NotEqual(t, r1, r2)
}

func makeHandler() rest.HandlerFunc {
	return func(r *http.Request) rest.RestResponse {
		s := fmt.Sprintf("%d", time.Now().UnixNano())
		return rest.NewResponse(s)
	}
}
