package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	authdomain "github.com/boreq/eggplant/domain/auth"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/eggplant/entrypoints/http/frontend"
	"github.com/boreq/eggplant/entrypoints/http/openapi"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
	"github.com/boreq/rest"
	"github.com/julienschmidt/httprouter"
)

var Version = "unknown"

type AuthProvider interface {
	Get(r *http.Request) (accessctx.AccessContext, error)
}

type Handler struct {
	app          *application.Application
	authProvider AuthProvider
	router       *httprouter.Router
	log          logging.Logger
}

func NewHandler(app *application.Application, authProvider AuthProvider) (*Handler, error) {
	h := &Handler{
		app:          app,
		authProvider: authProvider,
		router:       httprouter.New(),
		log:          logging.New("ports/http.Handler"),
	}

	// API
	h.router.HandlerFunc(http.MethodGet, "/api/browse", rest.Wrap(h.addAccessContextRest(h.browse)))
	h.router.HandlerFunc(http.MethodGet, "/api/browse/:id", rest.Wrap(h.addAccessContextRest(h.browseById)))
	h.router.HandlerFunc(http.MethodGet, "/api/stats", rest.Wrap(Cache(30*time.Second, h.stats)))
	h.router.HandlerFunc(http.MethodGet, "/api/search", rest.Wrap(h.addAccessContextRest(h.search)))

	h.router.HandlerFunc(http.MethodGet, "/api/track/:trackid/duration", rest.Wrap(h.addAccessContextRest(h.getTrackDuration)))
	h.router.HandlerFunc(http.MethodPost, "/api/track/:trackid/stream", rest.Wrap(h.addAccessContextRest(h.startTrackStream)))
	h.router.GET("/api/track/:trackid/stream/:streamid/playlist", h.addAccessContext(h.streamPlaylist))
	h.router.GET("/api/track/:trackid/stream/:streamid/init", h.addAccessContext(h.streamInit))
	h.router.GET("/api/track/:trackid/stream/:streamid/fragment/:number", h.addAccessContext(h.streamFragment))
	h.router.POST("/api/track/:trackid/stream/:streamid/keepalive", h.addAccessContext(h.keepAliveStream))
	h.router.GET("/api/thumbnail/:id", h.addAccessContext(h.thumbnail))

	h.router.HandlerFunc(http.MethodPost, "/api/auth/register-initial", rest.Wrap(h.registerInitial))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/register", rest.Wrap(h.addAccessContextRest(h.register)))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/login", rest.Wrap(h.addAccessContextRest(h.login)))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/logout", rest.Wrap(h.addAccessContextRest(h.logout)))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/create-invitation", rest.Wrap(h.addAccessContextRest(h.createInvitation)))
	h.router.HandlerFunc(http.MethodGet, "/api/auth", rest.Wrap(h.addAccessContextRest(h.getCurrentUser)))
	h.router.HandlerFunc(http.MethodGet, "/api/auth/users", rest.Wrap(h.addAccessContextRest(h.getUsers)))
	h.router.HandlerFunc(http.MethodGet, "/api/version", rest.Wrap(h.getVersion))
	h.router.HandlerFunc(http.MethodDelete, "/api/auth/users/:username", rest.Wrap(h.addAccessContextRest(h.removeUser)))

	// Remote instances: pairing administration (operator, admin only)
	h.router.HandlerFunc(http.MethodGet, "/api/remote", rest.Wrap(h.addAccessContextRest(h.remoteListRemotes)))
	h.router.HandlerFunc(http.MethodPost, "/api/remote", rest.Wrap(h.addAccessContextRest(h.remoteAddRemote)))
	h.router.HandlerFunc(http.MethodPost, "/api/remote/:id/pairing-token", rest.Wrap(h.addAccessContextRest(h.remoteSetPairingToken)))

	// Peer-to-peer endpoints
	h.router.HandlerFunc(http.MethodGet, "/api/peer/health", rest.Wrap(h.addAccessContextRest(h.remotePeerHealth)))
	h.router.HandlerFunc(http.MethodPost, "/api/peer/auth-token", rest.Wrap(h.remoteSetAuthToken))

	// API documentation
	h.router.HandlerFunc(http.MethodGet, "/api/openapi.yaml", h.openapiSpec)
	h.router.HandlerFunc(http.MethodGet, "/api", h.docs)

	// Frontend
	ffs, err := frontend.NewFrontendFileSystem()
	if err != nil {
		return nil, err
	}
	h.router.NotFound = http.FileServer(ffs)

	return h, nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.router.ServeHTTP(rw, req)
}

func (h *Handler) addAccessContextRest(handler func(accessctx.AccessContext, *http.Request) rest.RestResponse) rest.HandlerFunc {
	return func(r *http.Request) rest.RestResponse {
		accessCtx, err := h.authProvider.Get(r)
		if err != nil {
			if errors.Is(err, auth.ErrUnauthorized) {
				h.log.Debug("could not get the access context", "err", err)
				return rest.ErrUnauthorized
			}
			h.log.Error("could not get the access context", "err", err)
			return rest.ErrInternalServerError
		}
		return handler(accessCtx, r)
	}
}

func (h *Handler) addAccessContext(handler func(accessctx.AccessContext, http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		accessCtx, err := h.authProvider.Get(r)
		if err != nil {
			if errors.Is(err, auth.ErrUnauthorized) {
				h.log.Debug("could not get the access context", "err", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h.log.Error("could not get the access context", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		handler(accessCtx, w, r, ps)
	}
}

func (h *Handler) browse(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	a, err := h.app.Music.GetRootAlbum.Execute(accessCtx)
	if err != nil {
		return h.handleBrowseError(err)
	}
	return rest.NewResponse(toRootAlbum(a))
}

func (h *Handler) browseById(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	rawId := ps.ByName("id")

	albumId, err := musicdomain.NewAlbumIdFromString(rawId)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid album id.")
	}

	album, err := h.app.Music.GetAlbum.Execute(accessCtx, music.GetAlbum{Id: albumId})
	if err != nil {
		return h.handleBrowseError(err)
	}

	return rest.NewResponse(toAlbum(album))
}

func (h *Handler) handleBrowseError(err error) rest.RestResponse {
	if errors.Is(err, library.ErrAlbumNotFound) {
		return rest.ErrNotFound
	}
	if errors.Is(err, music.ErrLibraryNotReady) {
		return rest.ErrServiceUnavailable.WithMessage("The music library is being prepared.")
	}
	h.log.Error("browse error", "err", err)
	return rest.ErrInternalServerError
}

func (h *Handler) search(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	query, err := music.NewQuery(r.URL.Query().Get("query"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid query.")
	}

	result, err := h.app.Music.Search.Execute(accessCtx, music.Search{Query: query})
	if err != nil {
		h.log.Error("search error", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(
		toSearchResults(result),
	)
}

func (h *Handler) startTrackStream(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackid"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		return rest.ErrBadRequest.WithMessage("Invalid track id.")
	}

	seekPos, err := parseSeekParam(r.URL.Query().Get("seek"))
	if err != nil {
		h.log.Warn("invalid seek param", "err", err)
		return rest.ErrBadRequest.WithMessage("Invalid seek param.")
	}

	h.log.Debug("calling StartStreaming handler", "trackId", trackId.String(), "seekPos", seekPos)
	streamId, err := h.app.Music.StartStreaming.Execute(r.Context(), accessCtx, music.StartStreaming{
		TrackId:      trackId,
		SeekPosition: seekPos,
	})
	h.log.Debug("StartStreaming handler returned", "err", err)
	if err != nil {
		if errors.Is(err, library.ErrTrackNotFound) {
			return rest.ErrNotFound
		}
		if errors.Is(err, music.ErrTooManyOpenStreams) {
			return rest.ErrServiceUnavailable.WithMessage("Too many open streams.")
		}
		h.log.Error("start streaming failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(openapi.StartStreamResponse{StreamId: streamId.String()})
}

func (h *Handler) getTrackDuration(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())

	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackid"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		return rest.ErrBadRequest.WithMessage("Invalid track id.")
	}

	duration, err := h.app.Music.GetTrackDuration.Execute(r.Context(), accessCtx, music.GetTrackDuration{
		TrackId: trackId,
	})
	if err != nil {
		if errors.Is(err, library.ErrTrackNotFound) {
			return rest.ErrNotFound
		}
		if errors.Is(err, music.ErrLibraryNotReady) {
			return rest.ErrServiceUnavailable.WithMessage("The music library is being prepared.")
		}
		h.log.Error("get track duration failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(openapi.TrackDuration{Duration: duration.Seconds()})
}

func parseSeekParam(s string) (*musicdomain.RequestedSeekPosition, error) {
	if s == "" {
		return nil, nil
	}
	secs, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse seek seconds")
	}
	sp, err := musicdomain.NewRequestedSeekPosition(time.Duration(secs * float64(time.Second)))
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func (h *Handler) keepAliveStream(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, ok := h.parseStreamRequest(w, ps)
	if !ok {
		return
	}

	err := h.app.Music.KeepAliveStream.Execute(accessCtx, music.KeepAliveStream{
		TrackId:  trackId,
		StreamId: streamId,
	})
	if err != nil {
		h.translateAndWriteStreamError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) streamPlaylist(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, ok := h.parseStreamRequest(w, ps)
	if !ok {
		return
	}

	res, err := h.app.Music.StreamPlaylist.Execute(accessCtx, music.StreamPlaylist{
		TrackId:  trackId,
		StreamId: streamId,
	})
	if err != nil {
		h.translateAndWriteStreamError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	// While ffmpeg converts a track it rewrites the playlist several times per
	// second, each time with more segments. Last-Modified has 1s resolution, so
	// those sub-second updates collapse to one value and ServeContent answers
	// refreshes with spurious 304s: the client keeps a stale playlist that's
	// missing the later segments and stalls forever once it drains the ones it
	// has (issue #83). Pass a zero modtime to disable conditional requests;
	// no-store keeps it out of the browser cache too.
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "playlist.m3u8", time.Time{}, bytes.NewReader(res.Playlist.Bytes()))
}

func (h *Handler) streamInit(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, ok := h.parseStreamRequest(w, ps)
	if !ok {
		return
	}

	p, err := h.app.Music.StreamInit.Execute(accessCtx, music.StreamInit{
		TrackId:  trackId,
		StreamId: streamId,
	})
	if err != nil {
		h.translateAndWriteStreamError(w, err)
		return
	}
	defer p.Content.Close()

	h.serveConvertedFile(w, r, p, "video/mp4")
}

func (h *Handler) streamFragment(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, ok := h.parseStreamRequest(w, ps)
	if !ok {
		return
	}

	n, err := strconv.Atoi(ps.ByName("number"))
	if err != nil {
		h.log.Warn("invalid fragment number", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fragmentId, err := musicdomain.NewFragmentId(n)
	if err != nil {
		h.log.Warn("invalid fragment id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.app.Music.StreamFragment.Execute(accessCtx, music.StreamFragment{
		TrackId:    trackId,
		StreamId:   streamId,
		FragmentId: fragmentId,
	})
	if err != nil {
		h.translateAndWriteStreamError(w, err)
		return
	}
	defer p.Content.Close()

	h.serveConvertedFile(w, r, p, "video/iso.segment")
}

func (h *Handler) parseStreamRequest(w http.ResponseWriter, ps httprouter.Params) (musicdomain.TrackId, musicdomain.StreamId, bool) {
	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackid"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return musicdomain.TrackId{}, musicdomain.StreamId{}, false
	}
	streamId, err := musicdomain.NewStreamIdFromString(ps.ByName("streamid"))
	if err != nil {
		h.log.Warn("invalid stream id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return musicdomain.TrackId{}, musicdomain.StreamId{}, false
	}
	return trackId, streamId, true
}

var streamErrorMapping = []struct {
	err    error
	status int
}{
	{library.ErrTrackNotFound, http.StatusNotFound},
	{music.ErrStreamNotFound, http.StatusNotFound},
	{music.ErrStreamTrackMismatch, http.StatusNotFound},
	{music.ErrStreamPlaylistNotFound, http.StatusNotFound},
	{music.ErrStreamInitNotFound, http.StatusNotFound},
	{music.ErrStreamFragmentNotFound, http.StatusNotFound},
	{music.ErrLibraryNotReady, http.StatusServiceUnavailable},
}

func (h *Handler) translateAndWriteStreamError(w http.ResponseWriter, err error) {
	for _, m := range streamErrorMapping {
		if errors.Is(err, m.err) {
			w.WriteHeader(m.status)
			return
		}
	}
	h.log.Error("stream error", "err", err)
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *Handler) serveConvertedFile(w http.ResponseWriter, r *http.Request, p music.ConvertedFile, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeContent(w, r, "", p.Modtime, p.Content)
}

func (h *Handler) thumbnail(accessCtx accessctx.AccessContext, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := musicdomain.NewThumbnailIdFromString(ps.ByName("id"))
	if err != nil {
		h.log.Warn("invalid thumbnail id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.app.Music.Thumbnail.Execute(r.Context(), accessCtx, id)
	if err != nil {
		if errors.Is(err, library.ErrThumbnailNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("thumbnail error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer p.Content.Close()

	h.serveConvertedFile(w, r, p, "image/webp")
}

func (h *Handler) stats(r *http.Request) rest.RestResponse {
	stats, err := h.app.Queries.Stats.Execute()
	if err != nil {
		h.log.Error("stats query error", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(stats)
}

func (h *Handler) registerInitial(r *http.Request) rest.RestResponse {
	var t openapi.RegisterInitialInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("register initial decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	username, err := authdomain.NewUsernameFromString(t.Username)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid username.")
	}
	password, err := authdomain.NewPasswordFromString(t.Password)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid password.")
	}

	cmd := auth.RegisterInitial{
		Username: username,
		Password: password,
	}

	if err := h.app.Auth.RegisterInitial.Execute(cmd); err != nil {
		if errors.Is(err, auth.ErrUsersAlreadyExist) {
			return rest.ErrConflict.WithMessage("The initial setup has already been performed.")
		}
		h.log.Error("register initial command failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

func (h *Handler) login(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	var t openapi.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("login decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	username, err := authdomain.NewUsernameFromString(t.Username)
	if err != nil {
		return rest.ErrForbidden.WithMessage("Invalid credentials.")
	}

	password, err := authdomain.NewPasswordFromString(t.Password)
	if err != nil {
		return rest.ErrForbidden.WithMessage("Invalid credentials.")
	}

	cmd := auth.Login{
		Username: username,
		Password: password,
	}

	token, err := h.app.Auth.Login.Execute(accessCtx, cmd)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return rest.ErrForbidden.WithMessage("Invalid credentials.")
		}
		if errors.Is(err, auth.ErrAlreadyAuthenticated) {
			return rest.ErrBadRequest.WithMessage("You are already signed in.")
		}
		h.log.Error("login command failed", "err", err)
		return rest.ErrInternalServerError
	}

	response := openapi.LoginResult{
		Token: token.String(),
	}

	return rest.NewResponse(response)
}

func (h *Handler) logout(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	authCtx, ok := accessCtx.(accessctx.UserAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	if err := h.app.Auth.Logout.Execute(authCtx); err != nil {
		h.log.Error("could not logout the user", "err", err)
		return rest.ErrInternalServerError
	}
	return rest.NewResponse(nil)
}

func (h *Handler) getCurrentUser(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	authCtx, ok := accessCtx.(accessctx.UserAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	user, err := h.app.Auth.GetCurrentUser.Execute(authCtx)
	if err != nil {
		h.log.Error("could not get the current user", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(toReadUserResponse(user))
}

func (h *Handler) getUsers(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	users, err := h.app.Auth.List.Execute(accessCtx)
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can list users.")
		}
		h.log.Error("could not list", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(toReadUserResponses(users))
}

func (h *Handler) createInvitation(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	token, err := h.app.Auth.CreateInvitation.Execute(accessCtx)
	if err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can create invites.")
		}
		h.log.Error("could not create an invitation", "err", err)
		return rest.ErrInternalServerError
	}

	response := openapi.CreateInvitationResult{
		Token: token.String(),
	}

	return rest.NewResponse(response)
}

func (h *Handler) register(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	var t openapi.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("register decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	invitationToken, err := authdomain.NewInvitationTokenFromString(t.Token)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid invitation token.")
	}
	username, err := authdomain.NewUsernameFromString(t.Username)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid username.")
	}
	password, err := authdomain.NewPasswordFromString(t.Password)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid password.")
	}

	cmd := auth.Register{
		Username: username,
		Password: password,
		Token:    invitationToken,
	}

	if err := h.app.Auth.Register.Execute(accessCtx, cmd); err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			return rest.ErrConflict.WithMessage("Username is taken.")
		}
		if errors.Is(err, auth.ErrAlreadyAuthenticated) {
			return rest.ErrBadRequest.WithMessage("You are already signed in.")
		}
		h.log.Error("could not register a user", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

func (h *Handler) removeUser(accessCtx accessctx.AccessContext, r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	username, err := authdomain.NewUsernameFromString(ps.ByName("username"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid username.")
	}

	authCtx, ok := accessCtx.(accessctx.UserAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	cmd := auth.Remove{
		Username: username,
	}

	if err := h.app.Auth.Remove.Execute(authCtx, cmd); err != nil {
		if errors.Is(err, accessctx.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can remove users.")
		}
		if errors.Is(err, auth.ErrCannotRemoveSelf) {
			return rest.ErrBadRequest.WithMessage("You can not remove yourself.")
		}
		h.log.Error("could not remove a user", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

func (h *Handler) getVersion(r *http.Request) rest.RestResponse {
	return rest.NewResponse(openapi.VersionResponse{Version: Version})
}

func toReadUserResponse(u authdomain.User) openapi.ReadUserResponse {
	rv := openapi.ReadUserResponse{
		Username:      u.Username().String(),
		Administrator: u.Administrator(),
		Created:       u.Created(),
		LastSeen:      u.LastSeen(),
	}
	if sessions := u.Sessions(); len(sessions) > 0 {
		converted := make([]openapi.ReadSessionResponse, 0, len(sessions))
		for _, s := range sessions {
			converted = append(converted, openapi.ReadSessionResponse{LastSeen: s.LastSeen()})
		}
		rv.Sessions = &converted
	}
	return rv
}

func toReadUserResponses(users []authdomain.User) []openapi.ReadUserResponse {
	rv := make([]openapi.ReadUserResponse, 0, len(users))
	for _, u := range users {
		rv = append(rv, toReadUserResponse(u))
	}
	return rv
}
