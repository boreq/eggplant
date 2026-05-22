package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
	"github.com/boreq/eggplant/entrypoints/http/frontend"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
	"github.com/boreq/rest"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

var streamWsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const streamWsPingInterval = 15 * time.Second

var Version = "unknown"

type AuthenticatedUser struct {
	User  auth.ReadUser
	Token auth.AccessToken
}

type AuthProvider interface {
	Get(r *http.Request) (*AuthenticatedUser, error)
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

	h.router.GET("/api/track/:trackid/stream", h.trackStreamWS)
	h.router.GET("/api/track/:trackid/stream/:streamid/playlist", h.streamPlaylist)
	h.router.GET("/api/track/:trackid/stream/:streamid/init", h.streamInit)
	h.router.GET("/api/track/:trackid/stream/:streamid/fragment/:number", h.streamFragment)
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
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	accessCtx := accessContextFor(u)

	a, err := h.app.Music.GetRootAlbum.Execute(accessCtx)
	if err != nil {
		return h.handleBrowseError(err)
	}
	return rest.NewResponse(toRootAlbum(a))
}

func (h *Handler) browseById(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	rawId := ps.ByName("id")

	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	accessCtx := accessContextFor(u)

	albumId, err := domain.NewAlbumIdFromString(rawId)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid album id.")
	}

	a, err := h.app.Music.GetAlbum.Execute(accessCtx, music.GetAlbum{Id: albumId})
	if err != nil {
		return h.handleBrowseError(err)
	}

	return rest.NewResponse(toAlbum(a))
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
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	query, err := music.NewQuery(r.URL.Query().Get("query"))
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid query.")
	}

	result, err := h.app.Music.Search.Execute(accessContextFor(u), music.Search{Query: query})
	if err != nil {
		h.log.Error("search error", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(
		toSearchResult(result),
	)
}

func (h *Handler) trackStreamWS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, err := domain.NewTrackIdFromString(ps.ByName("trackid"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	seekPos, err := parseSeekParam(r.URL.Query().Get("seek"))
	if err != nil {
		h.log.Warn("invalid seek param", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessCtx, err := h.resolveAccessContext(r)
	if err != nil {
		h.log.Error("could not resolve access context", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, err := streamWsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("websocket upgrade failed", "err", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	streamId, err := h.app.Music.StartStreaming.Execute(ctx, accessCtx, music.StartStreaming{
		TrackId:      trackId,
		SeekPosition: seekPos,
	})
	if err != nil {
		if errors.Is(err, library.ErrTrackNotFound) {
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "track not found"))
			return
		}
		h.log.Error("start streaming failed", "err", err)
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "start streaming failed"))
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte(streamId.String())); err != nil {
		return
	}

	go h.keepWebsocketAlive(ctx, conn)

	// Hold the connection open
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.log.Debug("read error, exiting", "err", err)
			return
		}
	}
}

func (h *Handler) keepWebsocketAlive(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(streamWsPingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			deadline := time.Now().Add(streamWsPingInterval)
			if err := conn.WriteControl(websocket.PingMessage, nil, deadline); err != nil {
				h.log.Debug("ping error, exiting", "err", err)
				return
			}
		}
	}
}

func parseSeekParam(s string) (*domain.RequestedSeekPosition, error) {
	if s == "" {
		return nil, nil
	}
	secs, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse seek seconds")
	}
	sp, err := domain.NewRequestedSeekPosition(time.Duration(secs * float64(time.Second)))
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

func (h *Handler) streamPlaylist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackId, streamId, accessCtx, ok := h.parseStreamRequest(w, r, ps)
	if !ok {
		return
	}

	p, err := h.app.Music.StreamPlaylist.Execute(accessCtx, music.StreamPlaylist{
		TrackId:  trackId,
		StreamId: streamId,
	})
	if err != nil {
		h.writeStreamError(w, err)
		return
	}
	defer p.Content.Close()

	h.serveConvertedFile(w, r, p, "application/vnd.apple.mpegurl")
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
		h.writeStreamError(w, err)
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
	fragmentId, err := domain.NewFragmentId(n)
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
		h.writeStreamError(w, err)
		return
	}
	defer p.Content.Close()

	h.serveConvertedFile(w, r, p, "video/iso.segment")
}

func (h *Handler) parseStreamRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (domain.TrackId, domain.StreamId, library.AccessContext, bool) {
	trackId, err := domain.NewTrackIdFromString(ps.ByName("trackid"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return domain.TrackId{}, domain.StreamId{}, nil, false
	}
	streamId, err := domain.NewStreamIdFromString(ps.ByName("streamid"))
	if err != nil {
		h.log.Warn("invalid stream id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return domain.TrackId{}, domain.StreamId{}, nil, false
	}
	accessCtx, err := h.resolveAccessContext(r)
	if err != nil {
		h.log.Error("could not resolve access context", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return domain.TrackId{}, domain.StreamId{}, nil, false
	}
	return trackId, streamId, accessCtx, true
}

func (h *Handler) writeStreamError(w http.ResponseWriter, err error) {
	h.log.Warn("stream error", "err", err)
	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) resolveAccessContext(r *http.Request) (library.AccessContext, error) {
	u, err := h.authProvider.Get(r)
	if err != nil {
		return nil, errors.Wrap(err, "auth provider get failed")
	}
	return accessContextFor(u), nil
}

func (h *Handler) serveConvertedFile(w http.ResponseWriter, r *http.Request, p music.ConvertedFile, contentType string) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	http.ServeContent(w, r, p.Name, p.Modtime, p.Content)
}

func (h *Handler) thumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := domain.NewThumbnailIdFromString(ps.ByName("id"))
	if err != nil {
		h.log.Warn("invalid thumbnail id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessCtx, err := h.resolveAccessContext(r)
	if err != nil {
		h.log.Error("could not resolve access context", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
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

	cmd := auth.RegisterInitial{
		Username: t.Username,
		Password: t.Password,
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
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if u != nil {
		return rest.ErrBadRequest.WithMessage("You are already signed in.")
	}

	var t loginInput
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("login decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	cmd := auth.Login{
		Username: t.Username,
		Password: t.Password,
	}

	token, err := h.app.Auth.Login.Execute(cmd)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return rest.ErrForbidden.WithMessage("Invalid credentials.")
		}
		h.log.Error("login command failed", "err", err)
		return rest.ErrInternalServerError
	}

	response := loginResponse{
		Token: string(token),
	}

	return rest.NewResponse(response)
}

func (h *Handler) logout(r *http.Request) rest.RestResponse {
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if u == nil {
		return rest.ErrUnauthorized
	}

	cmd := auth.Logout{
		Token: u.Token,
	}

	if err := h.app.Auth.Logout.Execute(cmd); err != nil {
		h.log.Error("could not logout the user", "err", err)
		return rest.ErrInternalServerError
	}
	return rest.NewResponse(nil)
}

func (h *Handler) getCurrentUser(r *http.Request) rest.RestResponse {
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if u == nil {
		return rest.ErrUnauthorized
	}

	return rest.NewResponse(u.User)
}

func (h *Handler) getUsers(r *http.Request) rest.RestResponse {
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if !h.isAdmin(u) {
		return rest.ErrForbidden.WithMessage("Only an administrator can list users.")
	}

	users, err := h.app.Auth.List.Execute()
	if err != nil {
		h.log.Error("could not list", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(users)
}

type createInvitationResponse struct {
	Token string `json:"token"`
}

func (h *Handler) createInvitation(r *http.Request) rest.RestResponse {
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if !h.isAdmin(u) {
		return rest.ErrForbidden.WithMessage("Only an administrator can create invites.")
	}

	token, err := h.app.Auth.CreateInvitation.Execute()
	if err != nil {
		h.log.Error("could not create an invitation", "err", err)
		return rest.ErrInternalServerError
	}

	response := createInvitationResponse{
		Token: string(token),
	}

	return rest.NewResponse(response)
}

type registerInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (h *Handler) register(r *http.Request) rest.RestResponse {
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if u != nil {
		return rest.ErrBadRequest.WithMessage("You are signed in.")
	}

	var t registerInput
	if err = json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.log.Warn("register decoding failed", "err", err)
		return rest.ErrBadRequest.WithMessage("Malformed input.")
	}

	cmd := auth.Register{
		Username: t.Username,
		Password: t.Password,
		Token:    auth.InvitationToken(t.Token),
	}

	if err := h.app.Auth.Register.Execute(cmd); err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			return rest.ErrConflict.WithMessage("Username is taken.")
		}
		h.log.Error("could not register a user", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

func (h *Handler) removeUser(r *http.Request) rest.RestResponse {
	ps := httprouter.ParamsFromContext(r.Context())
	username := ps.ByName("username")

	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if !h.isAdmin(u) {
		return rest.ErrForbidden.WithMessage("Only an administrator can remove users.")
	}

	if username == u.User.Username {
		return rest.ErrBadRequest.WithMessage("You can not remove yourself.")
	}

	cmd := auth.Remove{
		Username: username,
	}

	if err := h.app.Auth.Remove.Execute(cmd); err != nil {
		h.log.Error("could not remove a user", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(nil)
}

type getVersionResponse struct {
	Version string `json:"version"`
}

func (h *Handler) getVersion(r *http.Request) rest.RestResponse {
	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	if !h.isAdmin(u) {
		return rest.ErrForbidden.WithMessage("Only an administrator can view the version.")
	}

	return rest.NewResponse(getVersionResponse{Version: Version})
}

func (h *Handler) isAdmin(u *AuthenticatedUser) bool {
	return u != nil && u.User.Administrator
}

func accessContextFor(u *AuthenticatedUser) library.AccessContext {
	if u == nil {
		return library.NewAnonymousAccessContext()
	}
	return library.NewLoggedInAccessContext()
}
