package remote

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/boreq/eggplant/adapters/openapi"
	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

const requestTimeout = 30 * time.Second

type RemoteClient struct {
	httpClient *http.Client
}

func NewRemoteClient() *RemoteClient {
	return &RemoteClient{
		httpClient: &http.Client{Timeout: requestTimeout},
	}
}

func (c *RemoteClient) SendLocalAuthToken(ctx context.Context, address remotedomain.RemoteInstanceAddress, remotePairingToken remotedomain.PairingToken, localAuthToken remotedomain.AuthToken) error {
	client, err := c.client(address)
	if err != nil {
		return err
	}

	// Field names are from the receiver's perspective: Token is the pairing
	// token the receiver gave us, OutboundToken is the token it uses to call back.
	body := openapi.SetRemoteAuthTokenJSONRequestBody{
		Token:         crockford.Encode(remotePairingToken.Bytes()),
		OutboundToken: crockford.Encode(localAuthToken.Bytes()),
	}

	resp, err := client.SetRemoteAuthTokenWithResponse(ctx, body)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	return expectOK(resp.StatusCode(), resp.Body)
}

func (c *RemoteClient) Healthcheck(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken) error {
	client, err := c.client(address)
	if err != nil {
		return err
	}

	resp, err := client.HealthWithResponse(ctx, authEditor(authToken))
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	return expectOK(resp.StatusCode(), resp.Body)
}

func (c *RemoteClient) GetRootAlbum(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken) (openapi.RootAlbum, error) {
	client, err := c.client(address)
	if err != nil {
		return openapi.RootAlbum{}, err
	}

	resp, err := client.BrowseWithResponse(ctx, authEditor(authToken))
	if err != nil {
		return openapi.RootAlbum{}, errors.Wrap(err, "request failed")
	}
	if resp.JSON200 == nil {
		return openapi.RootAlbum{}, errors.New(unexpectedStatus(resp.StatusCode(), resp.Body))
	}
	return *resp.JSON200, nil
}

func (c *RemoteClient) GetThumbnail(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, thumbnailId string) (io.ReadCloser, error) {
	client, err := openapi.NewClient(address.String(), openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not create the client")
	}
	resp, err := client.GetThumbnail(ctx, thumbnailId, authEditor(authToken))
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New(unexpectedStatus(resp.StatusCode, nil))
	}
	return resp.Body, nil
}

func (c *RemoteClient) GetTrackDuration(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, trackId string) (openapi.TrackDuration, error) {
	client, err := c.client(address)
	if err != nil {
		return openapi.TrackDuration{}, err
	}

	resp, err := client.GetTrackDurationWithResponse(ctx, trackId, authEditor(authToken))
	if err != nil {
		return openapi.TrackDuration{}, errors.Wrap(err, "request failed")
	}
	if resp.JSON200 == nil {
		return openapi.TrackDuration{}, errors.New(unexpectedStatus(resp.StatusCode(), resp.Body))
	}
	return *resp.JSON200, nil
}

func (c *RemoteClient) StartTrackStream(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, trackId string, seekSeconds *float64) (openapi.StartStreamResponse, error) {
	client, err := c.client(address)
	if err != nil {
		return openapi.StartStreamResponse{}, err
	}

	var params *openapi.StartTrackStreamParams
	if seekSeconds != nil {
		params = &openapi.StartTrackStreamParams{Seek: seekSeconds}
	}

	resp, err := client.StartTrackStreamWithResponse(ctx, trackId, params, authEditor(authToken))
	if err != nil {
		return openapi.StartStreamResponse{}, errors.Wrap(err, "request failed")
	}
	if resp.JSON200 == nil {
		return openapi.StartStreamResponse{}, errors.New(unexpectedStatus(resp.StatusCode(), resp.Body))
	}
	return *resp.JSON200, nil
}

func (c *RemoteClient) GetStreamPlaylist(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, trackId, streamId string) (io.ReadCloser, error) {
	client, err := openapi.NewClient(address.String(), openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not create the client")
	}
	resp, err := client.GetStreamPlaylist(ctx, trackId, streamId, authEditor(authToken))
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New(unexpectedStatus(resp.StatusCode, nil))
	}
	return resp.Body, nil
}

func (c *RemoteClient) GetStreamInit(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, trackId, streamId string) (io.ReadCloser, error) {
	client, err := openapi.NewClient(address.String(), openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not create the client")
	}
	resp, err := client.GetStreamInit(ctx, trackId, streamId, authEditor(authToken))
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New(unexpectedStatus(resp.StatusCode, nil))
	}
	return resp.Body, nil
}

func (c *RemoteClient) GetStreamFragment(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, trackId, streamId string, number int) (io.ReadCloser, error) {
	client, err := openapi.NewClient(address.String(), openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not create the client")
	}
	resp, err := client.GetStreamFragment(ctx, trackId, streamId, number, authEditor(authToken))
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errors.New(unexpectedStatus(resp.StatusCode, nil))
	}
	return resp.Body, nil
}

func (c *RemoteClient) KeepAliveStream(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, trackId, streamId string) error {
	client, err := openapi.NewClient(address.String(), openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return errors.Wrap(err, "could not create the client")
	}
	resp, err := client.KeepAliveStream(ctx, trackId, streamId, authEditor(authToken))
	if err != nil {
		return errors.Wrap(err, "request failed")
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(unexpectedStatus(resp.StatusCode, nil))
	}
	return nil
}

func (c *RemoteClient) GetAlbum(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, albumId string) (openapi.Album, error) {
	client, err := c.client(address)
	if err != nil {
		return openapi.Album{}, err
	}

	resp, err := client.BrowseAlbumWithResponse(ctx, albumId, authEditor(authToken))
	if err != nil {
		return openapi.Album{}, errors.Wrap(err, "request failed")
	}
	if resp.JSON200 == nil {
		return openapi.Album{}, errors.New(unexpectedStatus(resp.StatusCode(), resp.Body))
	}
	return *resp.JSON200, nil
}

func (c *RemoteClient) Search(ctx context.Context, address remotedomain.RemoteInstanceAddress, authToken remotedomain.AuthToken, query string) (openapi.SearchResults, error) {
	client, err := c.client(address)
	if err != nil {
		return openapi.SearchResults{}, err
	}

	params := &openapi.SearchParams{Query: query}
	resp, err := client.SearchWithResponse(ctx, params, authEditor(authToken))
	if err != nil {
		return openapi.SearchResults{}, errors.Wrap(err, "request failed")
	}
	if resp.JSON200 == nil {
		return openapi.SearchResults{}, errors.New(unexpectedStatus(resp.StatusCode(), resp.Body))
	}
	return *resp.JSON200, nil
}

func (c *RemoteClient) client(address remotedomain.RemoteInstanceAddress) (*openapi.ClientWithResponses, error) {
	client, err := openapi.NewClientWithResponses(address.String(), openapi.WithHTTPClient(c.httpClient))
	if err != nil {
		return nil, errors.Wrap(err, "could not create the client")
	}
	return client, nil
}

func authEditor(token remotedomain.AuthToken) openapi.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		SetAuthToken(req, token)
		return nil
	}
}

func expectOK(statusCode int, body []byte) error {
	if statusCode != http.StatusOK {
		return errors.New(unexpectedStatus(statusCode, body))
	}
	return nil
}

func unexpectedStatus(statusCode int, body []byte) string {
	if len(body) > 512 {
		body = body[:512]
	}
	return fmt.Sprintf("unexpected status code %d: %s", statusCode, string(body))
}
