package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

const requestTimeout = 30 * time.Second

// Field names are from the receiver's perspective.
type sendLocalAuthTokenRequest struct {
	Token         string `json:"token"`          // pairing token the receiver gave us
	OutboundToken string `json:"outbound_token"` // token the receiver uses to call back
}

type RemoteClient struct {
	httpClient *http.Client
}

func NewRemoteClient() *RemoteClient {
	return &RemoteClient{
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

func (c *RemoteClient) SendLocalAuthToken(ctx context.Context, address remotedomain.RemoteInstanceAddress, remotePairingToken remotedomain.PairingToken, localAuthToken remotedomain.AuthToken) error {
	body := sendLocalAuthTokenRequest{
		Token:         crockford.Encode(remotePairingToken.Bytes()),
		OutboundToken: crockford.Encode(localAuthToken.Bytes()),
	}

	j, err := json.Marshal(body)
	if err != nil {
		return errors.Wrap(err, "marshaling to json failed")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address.String()+"/api/peer/auth-token", bytes.NewReader(j))
	if err != nil {
		return errors.Wrap(err, "could not create the request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(unexpectedStatus(resp))
	}

	return nil
}

func (c *RemoteClient) Healthcheck(ctx context.Context, address remotedomain.RemoteInstanceAddress) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address.String()+"/api/peer/health", nil)
	if err != nil {
		return errors.Wrap(err, "could not create the request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(unexpectedStatus(resp))
	}

	return nil
}

func unexpectedStatus(resp *http.Response) string {
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Sprintf("unexpected status code %d: %s", resp.StatusCode, string(b))
}
