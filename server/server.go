// Package server exposes an HTTP API and serves the static files.
package server

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/server/api"
	_ "github.com/boreq/eggplant/statik"
	"github.com/boreq/eggplant/store"
	"github.com/julienschmidt/httprouter"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
)

var log = logging.New("server")

type handler struct {
	library        *library.Library
	trackStore     *store.TrackStore
	thumbnailStore *store.Store
}

func (h *handler) Browse(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
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

	d, err := h.library.Browse(ids)
	if err != nil {
		log.Error("browse error", "err", err)
		return nil, api.InternalServerError
	}

	return d, nil
}

var isIdValid = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func (h *handler) Track(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := trimExtension(ps.ByName("id"))
	if !isIdValid(id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	h.trackStore.ServeFile(w, r, id)
}

func (h *handler) Thumbnail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := trimExtension(ps.ByName("id"))
	if !isIdValid(id) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	h.thumbnailStore.ServeFile(w, r, id)
}

type Stats struct {
	Conversion ConversionStats `json:"conversion"`
}

type ConversionStats struct {
	Thumbnails store.Stats `json:"thumbnails"`
	Tracks     store.Stats `json:"tracks"`
}

func (h *handler) Stats(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	thumbnailStats, err := h.thumbnailStore.GetStats()
	if err != nil {
		log.Error("thumbnail stats error", "err", err)
		return nil, api.InternalServerError
	}

	trackStats, err := h.trackStore.GetStats()
	if err != nil {
		log.Error("track stats error", "err", err)
		return nil, api.InternalServerError
	}

	stats := Stats{
		Conversion: ConversionStats{
			Thumbnails: thumbnailStats,
			Tracks:     trackStats,
		},
	}
	return stats, nil
}

func trimExtension(s string) string {
	if index := strings.LastIndex(s, "."); index >= 0 {
		s = s[:index]
	}
	return s
}

func Serve(l *library.Library, trackStore *store.TrackStore, thumbnailStore *store.Store, address string) error {
	handler, err := newHandler(l, trackStore, thumbnailStore)
	if err != nil {
		return err
	}

	// Add CORS middleware
	handler = cors.AllowAll().Handler(handler)

	// Add GZIP middleware
	handler = gziphandler.GzipHandler(handler)

	log.Info("starting listening", "address", address)
	return http.ListenAndServe(address, handler)
}

type frontendFileSystem struct {
	fs http.FileSystem
}

func newFrontendFileSystem(fs http.FileSystem) *frontendFileSystem {
	return &frontendFileSystem{
		fs: fs,
	}
}

func (f *frontendFileSystem) Open(name string) (http.File, error) {
	file, err := f.fs.Open(name)
	if err != nil {
		file, err := f.fs.Open("/index.html")
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	return file, nil
}

func newHandler(l *library.Library, trackStore *store.TrackStore, thumbnailStore *store.Store) (http.Handler, error) {
	h := &handler{
		library:        l,
		trackStore:     trackStore,
		thumbnailStore: thumbnailStore,
	}

	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}

	router := httprouter.New()

	// API
	router.GET("/api/browse/*path", api.Wrap(h.Browse))
	router.GET("/api/track/:id", h.Track)
	router.GET("/api/thumbnail/:id", h.Thumbnail)
	router.GET("/api/stats", api.Wrap(h.Stats))

	// Frontend
	router.NotFound = http.FileServer(newFrontendFileSystem(statikFS))

	return router, nil
}
