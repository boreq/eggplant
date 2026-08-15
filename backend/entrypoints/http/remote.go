package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/boreq/eggplant/adapters/openapi"
	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/remote"
	"github.com/boreq/eggplant/domain/crockford"
	musicdomain "github.com/boreq/eggplant/domain/music"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
	"github.com/boreq/rest"
	"github.com/julienschmidt/httprouter"
)

type addRemoteInput struct {
	URL string `json:"url"`
}

func (h *Handler) remoteAddRemote(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
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

func (h *Handler) remoteListRemotes(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
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

func (h *Handler) remoteListRemoteLibraries(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	libraries, err := h.app.Remote.ListRemoteLibraries.Execute(accessCtx)
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden
		}
		h.log.Error("list remote libraries query failed", "err", err)
		return rest.ErrInternalServerError
	}

	out := make([]openapi.RemoteLibrary, 0, len(libraries))
	for _, library := range libraries {
		out = append(out, toRemoteLibrary(library))
	}

	return rest.NewResponse(out)
}

func toRemoteLibrary(library remote.RemoteLibrary) openapi.RemoteLibrary {
	return openapi.RemoteLibrary{
		Id:      library.ID().String(),
		Address: library.Address().String(),
	}
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

func (h *Handler) remotePeerHealth(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	if _, ok := accessCtx.(accessctx.RemoteInstanceAccessContext); !ok {
		return rest.ErrUnauthorized
	}

	return rest.NewResponse(openapi.PeerHealth{Service: remoteadapter.PeerServiceMarker})
}

func (h *Handler) remoteTrackDuration(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())

	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid remote instance id.")
	}

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackId"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid track id.")
	}

	duration, err := h.app.Music.RemoteGetTrackDuration.Execute(r.Context(), accessCtx, music.RemoteGetTrackDuration{InstanceId: instanceId, TrackId: trackId})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden
		}
		if errors.Is(err, remote.ErrNotFound) {
			return rest.ErrNotFound.WithMessage("Unknown remote instance.")
		}
		h.log.Error("could not get the remote track duration", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(openapi.TrackDuration{Duration: duration.Seconds()})
}

func (h *Handler) remoteRootAlbum(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())

	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid remote instance id.")
	}

	album, err := h.app.Music.RemoteGetRootAlbum.Execute(r.Context(), accessCtx, music.RemoteGetRootAlbum{InstanceId: instanceId})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden
		}
		if errors.Is(err, remote.ErrNotFound) {
			return rest.ErrNotFound.WithMessage("Unknown remote instance.")
		}
		h.log.Error("could not get the remote root album", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(toRootAlbum(album))
}

func (h *Handler) remoteAlbum(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())

	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid remote instance id.")
	}

	albumId, err := musicdomain.NewAlbumIdFromString(ps.ByName("albumId"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid album id.")
	}

	album, err := h.app.Music.RemoteGetAlbum.Execute(r.Context(), accessCtx, music.RemoteGetAlbum{InstanceId: instanceId, AlbumId: albumId})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden
		}
		if errors.Is(err, remote.ErrNotFound) {
			return rest.ErrNotFound.WithMessage("Unknown remote instance.")
		}
		h.log.Error("could not get the remote album", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(toAlbum(album))
}

func (h *Handler) remoteStartStream(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())

	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid remote instance id.")
	}

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackId"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid track id.")
	}

	seekPos, err := parseSeekParam(r.URL.Query().Get("seek"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid seek param.")
	}

	streamId, err := h.app.Music.RemoteStartStreaming.Execute(r.Context(), accessCtx, music.RemoteStartStreaming{
		InstanceId:   instanceId,
		TrackId:      trackId,
		SeekPosition: seekPos,
	})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden
		}
		if errors.Is(err, remote.ErrNotFound) {
			return rest.ErrNotFound.WithMessage("Unknown remote instance.")
		}
		h.log.Error("could not start the remote stream", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(openapi.StartStreamResponse{StreamId: streamId.String()})
}

func (h *Handler) remoteStreamPlaylist(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streamId, err := musicdomain.NewStreamIdFromString(ps.ByName("streamId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := h.app.Music.RemoteStreamPlaylist.Execute(r.Context(), accessCtx, music.RemoteStreamPlaylist{
		InstanceId: instanceId,
		TrackId:    trackId,
		StreamId:   streamId,
	})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, remote.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("could not get the remote stream playlist", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	if _, err := io.Copy(w, body); err != nil {
		h.log.Warn("remote stream playlist copy failed", "err", err)
	}
}

func (h *Handler) remoteStreamInit(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streamId, err := musicdomain.NewStreamIdFromString(ps.ByName("streamId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := h.app.Music.RemoteStreamInit.Execute(r.Context(), accessCtx, music.RemoteStreamInit{
		InstanceId: instanceId,
		TrackId:    trackId,
		StreamId:   streamId,
	})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, remote.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("could not get the remote stream init", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "video/mp4")
	if _, err := io.Copy(w, body); err != nil {
		h.log.Warn("remote stream init copy failed", "err", err)
	}
}

func (h *Handler) remoteStreamFragment(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streamId, err := musicdomain.NewStreamIdFromString(ps.ByName("streamId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(ps.ByName("number"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fragmentId, err := musicdomain.NewFragmentId(n)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := h.app.Music.RemoteStreamFragment.Execute(r.Context(), accessCtx, music.RemoteStreamFragment{
		InstanceId: instanceId,
		TrackId:    trackId,
		StreamId:   streamId,
		FragmentId: fragmentId,
	})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, remote.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("could not get the remote stream fragment", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "video/iso.segment")
	if _, err := io.Copy(w, body); err != nil {
		h.log.Warn("remote stream fragment copy failed", "err", err)
	}
}

func (h *Handler) remoteKeepAliveStream(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	streamId, err := musicdomain.NewStreamIdFromString(ps.ByName("streamId"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.app.Music.RemoteKeepAliveStream.Execute(r.Context(), accessCtx, music.RemoteKeepAliveStream{
		InstanceId: instanceId,
		TrackId:    trackId,
		StreamId:   streamId,
	}); err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, remote.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("could not keep alive the remote stream", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) remoteThumbnail(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	instanceId, err := remotedomain.NewRemoteInstanceIDFromString(ps.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	thumbnailId, err := musicdomain.NewThumbnailIdFromString(ps.ByName("thumbnailId"))
	if err != nil {
		h.log.Warn("invalid remote thumbnail id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := h.app.Music.RemoteGetThumbnail.Execute(r.Context(), accessCtx, music.RemoteGetThumbnail{
		InstanceId:  instanceId,
		ThumbnailId: thumbnailId,
	})
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, remote.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("could not get the remote thumbnail", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer body.Close()

	w.Header().Set("Content-Type", "image/webp")
	if _, err := io.Copy(w, body); err != nil {
		h.log.Warn("remote thumbnail copy failed", "err", err)
	}
}

type setRemotePairingTokenInput struct {
	PeerToken string `json:"peer_token"`
}

func (h *Handler) remoteSetPairingToken(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
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
