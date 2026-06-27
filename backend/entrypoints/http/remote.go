package http

import (
	"encoding/json"
	"net/http"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/entrypoints/http/openapi"
	"github.com/boreq/errors"
	"github.com/boreq/rest"
	"github.com/julienschmidt/httprouter"
)

type addRemoteInput struct {
	URL string `json:"url"`
}

func (h *Handler) remoteAddRemote(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	var t addRemoteInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("add remote decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	address, err := remotedomain.NewRemoteInstanceAddress(t.URL)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid address.")
	}

	result, err := h.app.Remote.AddRemote.Execute(accessCtx, remote.AddRemote{Address: address})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can manage remote instances.")
		}
		h.log.Error("add remote command failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(struct {
		ID                string `json:"id"`
		LocalPairingToken string `json:"local_pairing_token"`
	}{
		ID:                result.ID.String(),
		LocalPairingToken: crockford.Encode(result.LocalPairingToken.Bytes()),
	})
}

func (h *Handler) remoteListRemotes(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	instances, err := h.app.Remote.ListRemotes.Execute(accessCtx)
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can manage remote instances.")
		}
		h.log.Error("list remotes query failed", "err", err)
		return rest.ErrInternalServerError
	}

	out := make([]openapi.RemoteInstance, 0, len(instances))
	for _, instance := range instances {
		v, err := toRemoteInstance(instance)
		if err != nil {
			h.log.Error("could not convert the remote instance", "err", err)
			return rest.ErrInternalServerError
		}
		out = append(out, v)
	}

	return rest.NewResponse(out)
}

func toRemoteInstance(instance *remotedomain.RemoteInstance) (openapi.RemoteInstance, error) {
	status, err := toRemoteInstanceStatus(instance.Status())
	if err != nil {
		return openapi.RemoteInstance{}, errors.Wrap(err, "could not convert the status")
	}

	_, remotePairingTokenSet := instance.RemotePairingToken()

	rv := openapi.RemoteInstance{
		Id:                    instance.Id().String(),
		Address:               instance.Address().String(),
		Status:                status,
		RemotePairingTokenSet: remotePairingTokenSet,
	}
	if h, ok := instance.LastHealthcheck(); ok {
		status := openapi.RemoteInstanceLastHealthcheckStatus(h.Status().String())
		at := h.At()
		rv.LastHealthcheckStatus = &status
		rv.LastHealthcheckAt = &at
	}
	return rv, nil
}

func toRemoteInstanceStatus(status remotedomain.RemoteInstanceStatus) (openapi.RemoteInstanceStatus, error) {
	switch status {
	case remotedomain.RemoteInstanceStatusPairing:
		return openapi.RemoteInstanceStatusPAIRING, nil
	case remotedomain.RemoteInstanceStatusHealthy:
		return openapi.RemoteInstanceStatusHEALTHY, nil
	case remotedomain.RemoteInstanceStatusDead:
		return openapi.RemoteInstanceStatusDEAD, nil
	default:
		return openapi.RemoteInstanceStatus(""), errors.New("unknown remote instance status")
	}
}

func (h *Handler) remotePeerHealth(r *http.Request) rest.RestResponse {
	return rest.NewResponse(nil)
}

type setRemotePairingTokenInput struct {
	PeerToken string `json:"peer_token"`
}

func (h *Handler) remoteSetPairingToken(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	ps := httprouter.ParamsFromContext(r.Context())
	id, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid id.")
	}

	var t setRemotePairingTokenInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("set remote pairing token decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	peerToken, err := parsePairingToken(t.PeerToken)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid peer token.")
	}

	if err := h.app.Remote.SetRemotePairingToken.Execute(accessCtx, remote.SetRemotePairingToken{ID: id, RemotePairingToken: peerToken}); err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can manage remote instances.")
		}
		if errors.Is(err, remote.ErrNotFound) {
			return rest.ErrNotFound.WithMessage("Unknown remote instance.")
		}
		h.log.Error("set remote pairing token command failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

// Field names are from our (receiving) perspective.
type setRemoteAuthTokenInput struct {
	Token         string `json:"token"`
	OutboundToken string `json:"outbound_token"`
}

func (h *Handler) remoteSetAuthToken(r *http.Request) rest.RestResponse {
	var t setRemoteAuthTokenInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("set remote auth token decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	token, err := parsePairingToken(t.Token)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid token.")
	}

	outboundToken, err := parseAuthToken(t.OutboundToken)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid outbound token.")
	}

	cmd := remote.SetRemoteAuthToken{
		LocalPairingToken: token,
		RemoteAuthToken:   outboundToken,
	}

	if err := h.app.Remote.SetRemoteAuthToken.Execute(cmd); err != nil {
		if errors.Is(err, remote.ErrNotFound) {
			// Do not reveal whether the pairing token is known.
			return rest.ErrForbidden.WithMessage("Unknown pairing token.")
		}
		h.log.Error("set remote auth token command failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

func parsePairingToken(s string) (remotedomain.PairingToken, error) {
	b, err := crockford.Decode(s)
	if err != nil {
		return remotedomain.PairingToken{}, err
	}
	return remotedomain.NewPairingTokenFromBytes(b)
}

func parseAuthToken(s string) (remotedomain.AuthToken, error) {
	b, err := crockford.Decode(s)
	if err != nil {
		return remotedomain.AuthToken{}, err
	}
	return remotedomain.NewAuthTokenFromBytes(b)
}
