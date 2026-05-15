package http

import (
	"encoding/json"
	"net/http"
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
	"github.com/julienschmidt/httprouter"
)

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
	h.router.HandlerFunc(http.MethodGet, "/api/browse/:id", rest.Wrap(h.browse))
	h.router.HandlerFunc(http.MethodGet, "/api/stats", rest.Wrap(Cache(30*time.Second, h.stats)))
	h.router.HandlerFunc(http.MethodGet, "/api/search", rest.Wrap(h.search))

	h.router.GET("/api/track/:id", h.track)
	h.router.GET("/api/thumbnail/:id", h.thumbnail)

	h.router.HandlerFunc(http.MethodPost, "/api/auth/register-initial", rest.Wrap(h.registerInitial))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/register", rest.Wrap(h.register))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/login", rest.Wrap(h.login))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/logout", rest.Wrap(h.logout))
	h.router.HandlerFunc(http.MethodPost, "/api/auth/create-invitation", rest.Wrap(h.createInvitation))
	h.router.HandlerFunc(http.MethodGet, "/api/auth", rest.Wrap(h.getCurrentUser))
	h.router.HandlerFunc(http.MethodGet, "/api/auth/users", rest.Wrap(h.getUsers))
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
	ps := httprouter.ParamsFromContext(r.Context())
	rawId := ps.ByName("id")

	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		return rest.ErrInternalServerError
	}

	accessCtx := accessContextFor(u)

	if rawId == "" {
		a, err := h.app.Music.GetRootAlbum.Execute(accessCtx)
		if err != nil {
			if errors.Is(err, library.ErrAlbumNotFound) {
				return rest.ErrNotFound
			}
			h.log.Error("browse error", "err", err)
			return rest.ErrInternalServerError
		}
		return rest.NewResponse(toRootAlbum(a))
	}

	albumId, err := domain.NewAlbumIdFromString(rawId)
	if err != nil {
		return rest.ErrBadRequest.WithMessage("Invalid album id.")
	}

	a, err := h.app.Music.GetAlbum.Execute(accessCtx, music.GetAlbum{Id: albumId})
	if err != nil {
		if errors.Is(err, library.ErrAlbumNotFound) {
			return rest.ErrNotFound
		}
		h.log.Error("browse error", "err", err)
		return rest.ErrInternalServerError
	}

	return rest.NewResponse(toAlbum(a))
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

func (h *Handler) track(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := domain.NewTrackIdFromString(ps.ByName("id"))
	if err != nil {
		h.log.Warn("invalid track id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p, err := h.app.Music.Track.Execute(r.Context(), accessContextFor(u), id)
	if err != nil {
		if errors.Is(err, library.ErrTrackNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Error("track error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer p.Content.Close()

	w.Header().Add("Accept-Ranges", "bytes")
	http.ServeContent(w, r, p.Name, p.Modtime, p.Content)
}

func (h *Handler) thumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := domain.NewThumbnailIdFromString(ps.ByName("id"))
	if err != nil {
		h.log.Warn("invalid thumbnail id", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.authProvider.Get(r)
	if err != nil {
		h.log.Error("auth provider get failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	p, err := h.app.Music.Thumbnail.Execute(r.Context(), accessContextFor(u), id)
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

	w.Header().Add("Accept-Ranges", "bytes")
	http.ServeContent(w, r, p.Name, p.Modtime, p.Content)
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

func (h *Handler) isAdmin(u *AuthenticatedUser) bool {
	return u != nil && u.User.Administrator
}

func accessContextFor(u *AuthenticatedUser) library.AccessContext {
	if u == nil {
		return library.NewAnonymousAccessContext()
	}
	return library.NewLoggedInAccessContext()
}
