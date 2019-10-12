package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/eggplant/server/api"
	_ "github.com/boreq/eggplant/statik"
	"github.com/julienschmidt/httprouter"
	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
)

var log = logging.New("server")

type handler struct {
	library *library.Library
}

func (h *handler) Browse(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	path := ps.ByName("path")
	path = strings.Trim(path, "/")

	log.Debug("path", "path", path)

	var dirs []string
	if path != "" {
		dirs = strings.Split(path, "/")
	}

	var ids []library.Id
	for _, name := range dirs {
		ids = append(ids, library.Id(name))
	}

	d, err := h.library.Browse(ids)
	if err != nil {
		log.Error("browse error", "err", err)
		return nil, api.InternalServerError
	}

	return d, nil
}

func getParamInt(ps httprouter.Params, name string) (int, error) {
	return strconv.Atoi(getParamString(ps, name))
}

func getParamString(ps httprouter.Params, name string) string {
	return strings.TrimSuffix(ps.ByName(name), ".json")
}

func Serve(l *library.Library, address string) error {
	handler, err := newHandler(l)
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

func newHandler(l *library.Library) (http.Handler, error) {
	h := &handler{
		library: l,
	}

	statikFS, err := fs.New()
	if err != nil {
		return nil, err
	}

	router := httprouter.New()

	// API
	router.GET("/api/browse/*path", api.Wrap(h.Browse))

	// Frontend
	router.NotFound = http.FileServer(statikFS)

	return router, nil
}
