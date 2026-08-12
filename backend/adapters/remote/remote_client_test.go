package remote_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	testCases := []struct {
		name        string
		statusCode  int
		contentType string
		body        string
		expectedErr bool
	}{
		{
			name:        "peer_responds_with_the_marker",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body:        `{"service":"eggplant"}`,
			expectedErr: false,
		},
		{
			name:        "instance_without_the_endpoint_serves_the_frontend",
			statusCode:  http.StatusOK,
			contentType: "text/html; charset=utf-8",
			body:        "<!doctype html><html><body>eggplant</body></html>",
			expectedErr: true,
		},
		{
			name:        "some_other_service_responds_with_json",
			statusCode:  http.StatusOK,
			contentType: "application/json",
			body:        `{"service":"something-else"}`,
			expectedErr: true,
		},
		{
			name:        "peer_rejects_the_token",
			statusCode:  http.StatusUnauthorized,
			contentType: "application/json",
			body:        `{"message":"Unauthorized."}`,
			expectedErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var gotAuthorization string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuthorization = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", testCase.contentType)
				w.WriteHeader(testCase.statusCode)
				_, _ = w.Write([]byte(testCase.body))
			}))
			defer server.Close()

			address, err := remotedomain.NewRemoteInstanceAddress(server.URL)
			require.NoError(t, err)

			token, err := remotedomain.NewAuthToken()
			require.NoError(t, err)

			err = remoteadapter.NewRemoteClient().Healthcheck(context.Background(), address, token)
			if testCase.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NotEmpty(t, gotAuthorization)
		})
	}
}
