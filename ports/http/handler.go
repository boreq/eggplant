package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/ports/http/api"
	"github.com/boreq/eggplant/ports/http/frontend"
	"github.com/boreq/errors"
	"github.com/julienschmidt/httprouter"
)

var isIdValid = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

type Handler struct {
	app    *application.Application
	router *httprouter.Router
	log    logging.Logger
}

func NewHandler(app *application.Application) (*Handler, error) {
	h := &Handler{
		app:    app,
		router: httprouter.New(),
		log:    logging.New("ports/http.Handler"),
	}

	// API
	h.router.GET("/api/browse/*path", api.Wrap(h.browse))
	h.router.GET("/api/track/:id", h.track)
	h.router.GET("/api/thumbnail/:id", h.thumbnail)
	h.router.GET("/api/stats", api.Wrap(h.stats))

	h.router.POST("/api/auth/register-initial", api.Wrap(h.registerInitial))
	h.router.POST("/api/auth/register", api.Wrap(h.register))
	h.router.POST("/api/auth/login", api.Wrap(h.login))
	h.router.POST("/api/auth/logout", api.Wrap(h.logout))
	h.router.GET("/api/auth", api.Wrap(h.getCurrentUser))
	h.router.GET("/api/auth/users", api.Wrap(h.getUsers))
	h.router.POST("/api/auth/create-invitation", api.Wrap(h.createInvitation))

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

func (h *Handler) browse(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	u, err := h.getUser(r)
	if err != nil {
		h.log.Error("could not get the user", "err", err)
		return nil, api.InternalServerError
	}

	path := ps.ByName("path")
	path = strings.Trim(path, "/")

	var dirs []string
	if path != "" {
		dirs = strings.Split(path, "/")
	}

	var ids []music.AlbumId
	for _, name := range dirs {
		ids = append(ids, music.AlbumId(name))
	}

	cmd := music.Browse{
		Ids:        ids,
		PublicOnly: u == nil,
	}

	album, err := h.app.Music.Browse.Execute(cmd)
	if err != nil {
		h.log.Error("browse error", "err", err)
		return nil, api.InternalServerError
	}

	return album, nil
}

func (h *Handler) track(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if !isIdValid(id) {
		h.log.Warn("invalid track id", "id", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.app.Music.Track.Execute(id)
	if err != nil {
		h.log.Error("track error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	http.ServeFile(w, r, p)
}

func (h *Handler) thumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if !isIdValid(id) {
		h.log.Warn("invalid thumbnail id", "id", id)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	p, err := h.app.Music.Thumbnail.Execute(id)
	if err != nil {
		h.log.Error("track error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	http.ServeFile(w, r, p)
}

func (h *Handler) stats(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	stats, err := h.app.Queries.Stats.Execute()
	if err != nil {
		h.log.Error("stats query error", "err", err)
		return nil, api.InternalServerError
	}

	return stats, nil
}

type registerInitialInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) registerInitial(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	decoder := json.NewDecoder(r.Body)
	var t registerInitialInput
	err := decoder.Decode(&t)
	if err != nil {
		h.log.Warn("register initial decoding failed", "err", err)
		return nil, api.InternalServerError
	}

	cmd := auth.RegisterInitial{
		Username: t.Username,
		Password: t.Password,
	}

	if err := h.app.Auth.RegisterInitial.Execute(cmd); err != nil {
		h.log.Error("register initial command failed", "err", err)
		return nil, api.InternalServerError
	}

	return nil, nil
}

type loginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *Handler) login(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	decoder := json.NewDecoder(r.Body)
	var t loginInput
	if err := decoder.Decode(&t); err != nil {
		h.log.Warn("login decoding failed", "err", err)
		return nil, api.InternalServerError
	}

	cmd := auth.Login{
		Username: t.Username,
		Password: t.Password,
	}

	token, err := h.app.Auth.Login.Execute(cmd)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return nil, api.Forbidden
		}
		h.log.Error("initialize command failed", "err", err)
		return nil, api.InternalServerError
	}

	response := loginResponse{
		Token: string(token),
	}

	return response, nil
}

func (h *Handler) logout(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	u, err := h.getUser(r)
	if err != nil {
		h.log.Error("could not get the user", "err", err)
		return nil, api.InternalServerError
	}

	if u == nil {
		return nil, api.Unauthorized
	}

	token := h.getToken(r)

	cmd := auth.Logout{
		Token: auth.AccessToken(token),
	}

	if err := h.app.Auth.Logout.Execute(cmd); err != nil {
		h.log.Error("could not logout the user", "err", err)
		return nil, api.InternalServerError
	}
	return nil, nil
}

func (h *Handler) getCurrentUser(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	u, err := h.getUser(r)
	if err != nil {
		h.log.Error("could not the user", "err", err)
		return nil, api.InternalServerError
	}

	if u == nil {
		return nil, api.Unauthorized
	}

	return u, nil
}

func (h *Handler) getUsers(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	u, err := h.getUser(r)
	if err != nil {
		h.log.Error("could not the user", "err", err)
		return nil, api.InternalServerError
	}

	if !h.isAdmin(u) {
		return nil, api.Unauthorized
	}

	users, err := h.app.Auth.List.Execute()
	if err != nil {
		h.log.Error("could not list", "err", err)
		return nil, api.InternalServerError
	}

	return users, nil
}

type createInvitationResponse struct {
	Token string `json:"token"`
}

func (h *Handler) createInvitation(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	u, err := h.getUser(r)
	if err != nil {
		h.log.Error("could not the user", "err", err)
		return nil, api.InternalServerError
	}

	if !h.isAdmin(u) {
		return nil, api.Unauthorized
	}

	token, err := h.app.Auth.CreateInvitation.Execute()
	if err != nil {
		h.log.Error("could not list", "err", err)
		return nil, api.InternalServerError
	}

	response := createInvitationResponse{
		Token: string(token),
	}

	return response, nil
}

type registerInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func (h *Handler) register(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	u, err := h.getUser(r)
	if err != nil {
		h.log.Error("could not the user", "err", err)
		return nil, api.InternalServerError
	}

	if u != nil {
		return nil, api.BadRequest
	}

	decoder := json.NewDecoder(r.Body)
	var t registerInput
	if err = decoder.Decode(&t); err != nil {
		h.log.Warn("register decoding failed", "err", err)
		return nil, api.InternalServerError
	}

	cmd := auth.Register{
		Username: t.Username,
		Password: t.Password,
		Token:    auth.InvitationToken(t.Token),
	}

	if err := h.app.Auth.Register.Execute(cmd); err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			return nil, api.Conflict
		}
		h.log.Error("could not list", "err", err)
		return nil, api.InternalServerError
	}

	return nil, nil
}

func (h *Handler) isAdmin(u *auth.User) bool {
	return u != nil && u.Administrator
}

func (h *Handler) getUser(r *http.Request) (*auth.User, error) {
	token := h.getToken(r)
	if token == "" {
		return nil, nil
	}

	cmd := auth.CheckAccessToken{
		Token: auth.AccessToken(token),
	}

	user, err := h.app.Auth.CheckAccessToken.Execute(cmd)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "could not check the access token")
	}

	return &user, nil
}

func (h *Handler) getToken(r *http.Request) string {
	return r.Header.Get("Access-Token")
}
