package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	authdomain "github.com/boreq/eggplant/domain/auth"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/eggplant/entrypoints/http/frontend"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	"github.com/boreq/rest"
	"github.com/julienschmidt/httprouter"
)

type startStreamResponse struct {
	StreamId string `json:"streamId"`
}

var Version = "unknown"

type AuthProvider interface {
	Get(r *http.Request) (AccessContext, error)
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
	h.router.HandlerFunc(http.MethodGet, "/api/browse", rest.Wrap(h.browse))
	h.router.HandlerFunc(http.MethodGet, "/api/browse/:id", rest.Wrap(h.browseById))
	h.router.HandlerFunc(http.MethodGet, "/api/stats", rest.Wrap(Cache(30*time.Second, h.stats)))
	h.router.HandlerFunc(http.MethodGet, "/api/search", rest.Wrap(h.search))

	h.router.HandlerFunc(http.MethodPost, "/api/track/:trackid/stream", rest.Wrap(h.startTrackStream))
	h.router.GET("/api/track/:trackid/stream/:streamid/playlist", h.streamPlaylist)
	h.router.GET("/api/track/:trackid/stream/:streamid/init", h.streamInit)
	h.router.GET("/api/track/:trackid/stream/:streamid/fragment/:number", h.streamFragment)
	h.router.POST("/api/track/:trackid/stream/:streamid/keepalive", h.keepAliveStream)
	h.router.GET("/api/thumbnail/:id", h.thumbnail)

	h.router.HandlerFunc(http.MethodPost, "/api/auth/register-initial", rest.Wrap(h.registerInitial))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/register", rest.Wrap(h.register))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/login", rest.Wrap(h.login))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/logout", rest.Wrap(h.logout))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/create-invitation", rest.Wrap(h.createInvitation))
	h.router.HandlerFunc(http.MethodGet, "/api/auth", rest.Wrap(h.getCurrentUser))
	h.router.HandlerFunc(http.MethodGet, "/api/auth/users", rest.Wrap(h.getUsers))
	h.router.HandlerFunc(http.MethodGet, "/api/version", rest.Wrap(h.getVersion))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/users/:username/remove", rest.Wrap(h.removeUser))

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

func (h *Handler) browse(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	a, err := h.app.Music.GetRootAlbum.Execute(accessCtx.Library())
	if err != nil {
		return h.handleBrowseError(err)
	}
	return rest.NewResponse(toRootAlbum(a))
}

func (h *Handler) browseById(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	rawId := ps.ByName("id")

	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	albumId, err := musicdomain.NewAlbumIdFromString(rawId)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid album id.")
	}

	album, err := h.app.Music.GetAlbum.Execute(accessCtx.Library(), music.GetAlbum{Id: albumId})
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

func (h *Handler) search(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	query, err := music.NewQuery(r.URL.Query().Get("query"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid query.")
	}

	result, err := h.app.Music.Search.Execute(accessCtx.Library(), music.Search{Query: query})
	if err != nil {
		h.log.Error("search error", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(
		toSearchResults(result),
	)
}

func (h *Handler) startTrackStream(r *http.Request) rest.RestResponse {
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

	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	streamId, err := h.app.Music.StartStreaming.Execute(r.Context(), accessCtx.Library(), music.StartStreaming{
		TrackId:      trackId,
		SeekPosition: seekPos,
	})
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

	return rest.NewResponse(startStreamResponse{StreamId: streamId.String()})
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

func (h *Handler) keepAliveStream(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, accessCtx, ok := h.parseStreamRequest(w, r, ps)
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

func (h *Handler) streamPlaylist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, accessCtx, ok := h.parseStreamRequest(w, r, ps)
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
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeContent(w, r, "playlist.m3u8", res.Modtime, bytes.NewReader(res.Playlist.Bytes()))
}

func (h *Handler) streamInit(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, accessCtx, ok := h.parseStreamRequest(w, r, ps)
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

func (h *Handler) streamFragment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, accessCtx, ok := h.parseStreamRequest(w, r, ps)
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

func (h *Handler) parseStreamRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (musicdomain.TrackId, musicdomain.StreamId, library.AccessContext, bool) {
	trackId, err := musicdomain.NewTrackIdFromString(ps.ByName("trackid"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return musicdomain.TrackId{}, musicdomain.StreamId{}, nil, false
	}
	streamId, err := musicdomain.NewStreamIdFromString(ps.ByName("streamid"))
	if err != nil {
		h.log.Warn("invalid stream id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return musicdomain.TrackId{}, musicdomain.StreamId{}, nil, false
	}
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return musicdomain.TrackId{}, musicdomain.StreamId{}, nil, false
	}
	return trackId, streamId, accessCtx.Library(), true
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

func (h *Handler) thumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := musicdomain.NewThumbnailIdFromString(ps.ByName("id"))
	if err != nil {
		h.log.Warn("invalid thumbnail id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p, err := h.app.Music.Thumbnail.Execute(r.Context(), accessCtx.Library(), id)
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

type registerInitialInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) registerInitial(r *http.Request) rest.RestResponse {
	var t registerInitialInput
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
		h.log.Error("register initial command failed", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

type loginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) login(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	var t loginInput
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

	token, err := h.app.Auth.Login.Execute(accessCtx.Auth(), cmd)
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

	response := loginResponse{
		Token: token.String(),
	}

	return rest.NewResponse(response)
}

func (h *Handler) logout(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	if err := h.app.Auth.Logout.Execute(authCtx); err != nil {
		h.log.Error("could not logout the user", "err", err)
		return rest.ErrInternalServerError
	}
	return rest.NewResponse(nil)
}

func (h *Handler) getCurrentUser(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
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

func (h *Handler) getUsers(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	users, err := h.app.Auth.List.Execute(accessCtx.Auth())
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can list users.")
		}
		h.log.Error("could not list", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(toReadUserResponses(users))
}

type createInvitationResponse struct {
	Token string `json:"token"`
}

func (h *Handler) createInvitation(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	token, err := h.app.Auth.CreateInvitation.Execute(accessCtx.Auth())
	if err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
			return rest.ErrForbidden.WithMessage("Only an administrator can create invites.")
		}
		h.log.Error("could not create an invitation", "err", err)
		return rest.ErrInternalServerError
	}

	response := createInvitationResponse{
		Token: token.String(),
	}

	return rest.NewResponse(response)
}

type registerInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (h *Handler) register(r *http.Request) rest.RestResponse {
	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	var t registerInput
	if err = json.NewDecoder(r.Body).Decode(&t); err != nil {
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

	if err := h.app.Auth.Register.Execute(accessCtx.Auth(), cmd); err != nil {
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

func (h *Handler) removeUser(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	username, err := authdomain.NewUsernameFromString(ps.ByName("username"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid username.")
	}

	accessCtx, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	authCtx, ok := accessCtx.Auth().(auth.AuthenticatedAccessContext)
	if !ok {
		return rest.ErrUnauthorized
	}

	cmd := auth.Remove{
		Username: username,
	}

	if err := h.app.Auth.Remove.Execute(authCtx, cmd); err != nil {
		if errors.Is(err, auth.ErrPermissionDenied) {
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

type getVersionResponse struct {
	Version string `json:"version"`
}

func (h *Handler) getVersion(r *http.Request) rest.RestResponse {
	return rest.NewResponse(getVersionResponse{Version: Version})
}

type readSessionResponse struct {
	LastSeen time.Time `json:"lastSeen"`
}

type readUserResponse struct {
	Username      string                `json:"username"`
	Administrator bool                  `json:"administrator"`
	Created       time.Time             `json:"created"`
	LastSeen      time.Time             `json:"lastSeen"`
	Sessions      []readSessionResponse `json:"sessions"`
}

func toReadUserResponse(u authdomain.User) readUserResponse {
	rv := readUserResponse{
		Username:      u.Username().String(),
		Administrator: u.Administrator(),
		Created:       u.Created(),
		LastSeen:      u.LastSeen(),
	}
	for _, s := range u.Sessions() {
		rv.Sessions = append(rv.Sessions, readSessionResponse{LastSeen: s.LastSeen()})
	}
	return rv
}

func toReadUserResponses(users []authdomain.User) []readUserResponse {
	rv := make([]readUserResponse, 0, len(users))
	for _, u := range users {
		rv = append(rv, toReadUserResponse(u))
	}
	return rv
}
