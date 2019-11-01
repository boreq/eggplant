package http

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/boreq/eggplant/errors"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/pkg/service/application"
	"github.com/boreq/eggplant/pkg/service/application/auth"
	"github.com/boreq/eggplant/pkg/service/ports/http/api"
	_ "github.com/boreq/eggplant/statik"
	"github.com/boreq/eggplant/store"
	"github.com/julienschmidt/httprouter"
	"github.com/rakyll/statik/fs"
)

var isIdValid = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

type Stats struct {
	Conversion ConversionStats `json:"conversion"`
	Users      int             `json:"users"`
}

type ConversionStats struct {
	Thumbnails store.Stats `json:"thumbnails"`
	Tracks     store.Stats `json:"tracks"`
}

type Handler struct {
	app            *application.Application
	lib            *library.Library
	trackStore     *store.TrackStore
	thumbnailStore *store.Store

	router *httprouter.Router
	log    logging.Logger
}

func NewHandler(app *application.Application, lib *library.Library, trackStore *store.TrackStore, thumbnailStore *store.Store) (*Handler, error) {
	h := &Handler{
		app:            app,
		lib:            lib,
		trackStore:     trackStore,
		thumbnailStore: thumbnailStore,
		router:         httprouter.New(),
		log:            logging.New("ports/http.Handler"),
	}

	// API
	h.router.GET("/api/browse/*path", api.Wrap(h.browse))
	h.router.GET("/api/track/:id", h.track)
	h.router.GET("/api/thumbnail/:id", h.thumbnail)
	h.router.GET("/api/stats", api.Wrap(h.stats))

	h.router.POST("/api/user/register-initial", api.Wrap(h.registerInitial))
	h.router.POST("/api/user/login", api.Wrap(h.login))
	h.router.POST("/api/user/logout", api.Wrap(h.logout))
	h.router.GET("/api/user", api.Wrap(h.getCurrentUser))

	// Frontend
	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}
	h.router.NotFound = http.FileServer(newFrontendFileSystem(statikFS))

	return h, nil
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.router.ServeHTTP(rw, req)
}

func (h *Handler) browse(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	path := ps.ByName("path")
	path = strings.Trim(path, "/")

	var dirs []string
	if path != "" {
		dirs = strings.Split(path, "/")
	}

	var ids []library.AlbumId
	for _, name := range dirs {
		ids = append(ids, library.AlbumId(name))
	}

	d, err := h.lib.Browse(ids)
	if err != nil {
		h.log.Error("browse error", "err", err)
		return nil, api.InternalServerError
	}

	return d, nil
}

func (h *Handler) track(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if !isIdValid(id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	h.trackStore.ServeFile(w, r, id)
}

func (h *Handler) thumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	if !isIdValid(id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	h.thumbnailStore.ServeFile(w, r, id)
}

func (h *Handler) stats(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	thumbnailStats, err := h.thumbnailStore.GetStats()
	if err != nil {
		h.log.Error("thumbnail stats error", "err", err)
		return nil, api.InternalServerError
	}

	trackStats, err := h.trackStore.GetStats()
	if err != nil {
		h.log.Error("track stats error", "err", err)
		return nil, api.InternalServerError
	}

	st, err := h.app.Queries.Stats.Execute()
	if err != nil {
		h.log.Error("stats query error", "err", err)
		return nil, api.InternalServerError
	}

	stats := Stats{
		Conversion: ConversionStats{
			Thumbnails: thumbnailStats,
			Tracks:     trackStats,
		},
		Users: st.Users,
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
