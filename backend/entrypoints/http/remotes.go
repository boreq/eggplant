package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/boreq/eggplant/adapters/remotes"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/errors"
	"github.com/boreq/rest"
	"github.com/julienschmidt/httprouter"
)

// RemoteRepository is the interface the handler uses to load/store remote instances.
type RemoteRepository interface {
	List(username string) ([]remotes.RemoteInstance, error)
	Get(username, id string) (*remotes.RemoteInstance, error)
	Put(username string, inst remotes.RemoteInstance) (remotes.RemoteInstance, error)
	Delete(username, id string) error
}

type addRemoteInput struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type remoteInstanceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// RemoteClient authenticates against remote instances and proxies requests.
// It caches auth tokens in memory; on 401 it re-authenticates and retries.
type RemoteClient struct {
	mu     sync.Mutex
	tokens map[string]string // remoteID -> token
}

func NewRemoteClient() *RemoteClient {
	return &RemoteClient{tokens: make(map[string]string)}
}

func (c *RemoteClient) do(ctx context.Context, inst *remotes.RemoteInstance, method, url string, body io.Reader) (*http.Response, error) {
	token, err := c.getToken(ctx, inst)
	if err != nil {
		return nil, errors.Wrap(err, "could not get auth token for remote")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "auth-token", Value: token})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.mu.Lock()
		delete(c.tokens, inst.ID)
		c.mu.Unlock()

		token, err = c.getToken(ctx, inst)
		if err != nil {
			return nil, errors.Wrap(err, "could not re-authenticate with remote")
		}
		req2, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		req2.AddCookie(&http.Cookie{Name: "auth-token", Value: token})
		return http.DefaultClient.Do(req2)
	}

	return resp, nil
}

func (c *RemoteClient) getToken(ctx context.Context, inst *remotes.RemoteInstance) (string, error) {
	c.mu.Lock()
	token, ok := c.tokens[inst.ID]
	c.mu.Unlock()
	if ok {
		return token, nil
	}

	token, err := c.login(ctx, inst)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.tokens[inst.ID] = token
	c.mu.Unlock()

	return token, nil
}

func (c *RemoteClient) login(ctx context.Context, inst *remotes.RemoteInstance) (string, error) {
	body, err := json.Marshal(map[string]string{
		"username": inst.Username,
		"password": inst.Password,
	})
	if err != nil {
		return "", err
	}

	loginURL := strings.TrimRight(inst.URL, "/") + "/api/auth/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return "", errors.Wrap(err, "could not create login request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "login request to remote failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote login returned status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.Wrap(err, "could not decode login response")
	}
	if result.Token == "" {
		return "", errors.New("remote returned empty token")
	}
	return result.Token, nil
}

func (c *RemoteClient) InvalidateToken(remoteID string) {
	c.mu.Lock()
	delete(c.tokens, remoteID)
	c.mu.Unlock()
}

// --- Management endpoints ---

func (h *Handler) listRemotes(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		return rest.ErrInternalServerError
	}
	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	instances, err := h.remoteRepo.List(authCtx.Username().String())
	if err != nil {
		h.log.Error("could not list remotes", "err", err)
		return rest.ErrInternalServerError
	}

	result := make([]remoteInstanceResponse, 0, len(instances))
	for _, inst := range instances {
		result = append(result, remoteInstanceResponse{ID: inst.ID, Name: inst.Name, URL: inst.URL})
	}
	return rest.NewResponse(result)
}

func (h *Handler) addRemote(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		return rest.ErrInternalServerError
	}
	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	var input addRemoteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}
	if input.Name == "" || input.URL == "" || input.Username == "" || input.Password == "" {
		return rest.ErrBadRequest.WithMessage("name, url, username, and password are required.")
	}

	inst, err := h.remoteRepo.Put(authCtx.Username().String(), remotes.RemoteInstance{
		Name:     input.Name,
		URL:      strings.TrimRight(input.URL, "/"),
		Username: input.Username,
		Password: input.Password,
	})
	if err != nil {
		h.log.Error("could not add remote", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(remoteInstanceResponse{ID: inst.ID, Name: inst.Name, URL: inst.URL})
}

func (h *Handler) removeRemote(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		return rest.ErrInternalServerError
	}
	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	ps := httprouter.ParamsFromContext(r.Context())
	remoteID := ps.ByName("remoteid")
	if remoteID == "" {
		return rest.ErrBadRequest.WithMessage("Missing remote id.")
	}

	if err := h.remoteRepo.Delete(authCtx.Username().String(), remoteID); err != nil {
		h.log.Error("could not delete remote", "err", err)
		return rest.ErrInternalServerError
	}

	h.remoteClient.InvalidateToken(remoteID)
	return rest.NewResponse(nil)
}

// --- Helpers to resolve the remote instance from a request ---

func (h *Handler) getRemoteInstanceForRequest(r *http.Request, remoteID string) (*remotes.RemoteInstance, rest.RestResponse) {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		return nil, rest.ErrInternalServerError
	}
	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
	if !ok {
		return nil, rest.ErrUnauthorized
	}

	inst, err := h.remoteRepo.Get(authCtx.Username().String(), remoteID)
	if err != nil {
		h.log.Error("could not get remote instance", "err", err)
		return nil, rest.ErrInternalServerError
	}
	if inst == nil {
		return nil, rest.ErrNotFound
	}
	return inst, nil
}

// --- Proxy endpoints ---

func (h *Handler) remoteBrowseRoot(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		return errResp
	}
	return h.proxyJSONGet(r, inst, inst.URL+"/api/browse")
}

func (h *Handler) remoteBrowseByID(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		return errResp
	}
	return h.proxyJSONGet(r, inst, inst.URL+"/api/browse/"+ps.ByName("albumid"))
}

func (h *Handler) remoteSearch(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		return errResp
	}
	return h.proxyJSONGet(r, inst, inst.URL+"/api/search?query="+r.URL.Query().Get("query"))
}

func (h *Handler) remoteTrackDuration(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		return errResp
	}
	return h.proxyJSONGet(r, inst, inst.URL+"/api/track/"+ps.ByName("trackid")+"/duration")
}

func (h *Handler) remoteStartStream(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		return errResp
	}
	remoteURL := inst.URL + "/api/track/" + ps.ByName("trackid") + "/stream"
	if seek := r.URL.Query().Get("seek"); seek != "" {
		remoteURL += "?seek=" + seek
	}
	return h.proxyJSONDo(r, inst, http.MethodPost, remoteURL, nil)
}

func (h *Handler) remoteStreamPlaylist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		w.WriteHeader(errResp.StatusCode())
		return
	}
	remoteURL := inst.URL + "/api/track/" + ps.ByName("trackid") +
		"/stream/" + ps.ByName("streamid") + "/playlist"
	h.proxyBinary(w, r, inst, remoteURL)
}

func (h *Handler) remoteStreamInit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		w.WriteHeader(errResp.StatusCode())
		return
	}
	remoteURL := inst.URL + "/api/track/" + ps.ByName("trackid") +
		"/stream/" + ps.ByName("streamid") + "/init"
	h.proxyBinary(w, r, inst, remoteURL)
}

func (h *Handler) remoteStreamFragment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		w.WriteHeader(errResp.StatusCode())
		return
	}
	remoteURL := inst.URL + "/api/track/" + ps.ByName("trackid") +
		"/stream/" + ps.ByName("streamid") + "/fragment/" + ps.ByName("number")
	h.proxyBinary(w, r, inst, remoteURL)
}

func (h *Handler) remoteKeepAlive(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		w.WriteHeader(errResp.StatusCode())
		return
	}
	remoteURL := inst.URL + "/api/track/" + ps.ByName("trackid") +
		"/stream/" + ps.ByName("streamid") + "/keepalive"
	resp, err := h.remoteClient.do(r.Context(), inst, http.MethodPost, remoteURL, nil)
	if err != nil {
		h.log.Error("remote keepalive failed", "err", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
}

func (h *Handler) remoteThumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	inst, errResp := h.getRemoteInstanceForRequest(r, ps.ByName("remoteid"))
	if errResp != nil {
		w.WriteHeader(errResp.StatusCode())
		return
	}
	h.proxyBinary(w, r, inst, inst.URL+"/api/thumbnail/"+ps.ByName("id"))
}

// proxyJSONGet proxies a GET request and forwards the JSON response.
func (h *Handler) proxyJSONGet(r *http.Request, inst *remotes.RemoteInstance, remoteURL string) rest.RestResponse {
	return h.proxyJSONDo(r, inst, http.MethodGet, remoteURL, nil)
}

// proxyJSONDo proxies any method and forwards the JSON response.
func (h *Handler) proxyJSONDo(r *http.Request, inst *remotes.RemoteInstance, method, remoteURL string, body io.Reader) rest.RestResponse {
	resp, err := h.remoteClient.do(r.Context(), inst, method, remoteURL, body)
	if err != nil {
		h.log.Error("remote request failed", "url", remoteURL, "err", err)
		return rest.ErrBadGateway
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return rest.NewError(resp.StatusCode, "remote error")
	}

	var payload json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		h.log.Error("could not decode remote response", "err", err)
		return rest.ErrInternalServerError
	}
	return rest.NewResponse(payload)
}

// proxyBinary proxies a request forwarding binary response data directly.
func (h *Handler) proxyBinary(w http.ResponseWriter, r *http.Request, inst *remotes.RemoteInstance, remoteURL string) {
	resp, err := h.remoteClient.do(r.Context(), inst, http.MethodGet, remoteURL, nil)
	if err != nil {
		h.log.Error("remote binary request failed", "url", remoteURL, "err", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for _, hdr := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Cache-Control"} {
		if v := resp.Header.Get(hdr); v != "" {
			w.Header().Set(hdr, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) //nolint:errcheck
}
