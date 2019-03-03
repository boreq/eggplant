package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"

	"github.com/boreq/goaccess/core"
	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/server/api"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

var log = logging.New("server")

type handler struct {
	repository *core.Repository
}

func (h *handler) Hour(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	year, err := getParamInt(ps, "year")
	if err != nil {
		return nil, api.BadRequest
	}

	month, err := getParamInt(ps, "month")
	if err != nil {
		return nil, api.BadRequest
	}

	day, err := getParamInt(ps, "day")
	if err != nil {
		return nil, api.BadRequest
	}

	hour, err := getParamInt(ps, "hour")
	if err != nil {
		return nil, api.BadRequest
	}

	if month < 1 || month > 12 {
		return nil, api.BadRequest
	}

	data, ok := h.repository.RetrieveHour(year, time.Month(month), day, hour)
	if !ok {
		return nil, api.NotFound
	}
	rangeData := RangeData{
		Time: time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC),
		Data: data,
	}
	return rangeData, nil
}

func (h *handler) Day(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	year, err := getParamInt(ps, "year")
	if err != nil {
		return nil, api.BadRequest
	}

	month, err := getParamInt(ps, "month")
	if err != nil {
		return nil, api.BadRequest
	}

	day, err := getParamInt(ps, "day")
	if err != nil {
		return nil, api.BadRequest
	}

	if month < 1 || month > 12 {
		return nil, api.BadRequest
	}

	data, ok := h.repository.RetrieveDay(year, time.Month(month), day)
	if !ok {
		return nil, api.NotFound
	}
	rangeData := RangeData{
		Time: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
		Data: data,
	}
	return rangeData, nil
}

func (h *handler) Month(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	year, err := getParamInt(ps, "year")
	if err != nil {
		return nil, api.BadRequest
	}

	month, err := getParamInt(ps, "month")
	if err != nil {
		return nil, api.BadRequest
	}

	if month < 1 || month > 12 {
		return nil, api.BadRequest
	}

	data, ok := h.repository.RetrieveMonth(year, time.Month(month))
	if !ok {
		return nil, api.NotFound
	}
	rangeData := RangeData{
		Time: time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC),
		Data: data,
	}
	return rangeData, nil
}

func (h *handler) RangeHourly(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	yearFrom, err := getParamInt(ps, "yearFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	monthFrom, err := getParamInt(ps, "monthFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	dayFrom, err := getParamInt(ps, "dayFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	hourFrom, err := getParamInt(ps, "hourFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	yearTo, err := getParamInt(ps, "yearTo")
	if err != nil {
		return nil, api.BadRequest
	}

	monthTo, err := getParamInt(ps, "monthTo")
	if err != nil {
		return nil, api.BadRequest
	}

	dayTo, err := getParamInt(ps, "dayTo")
	if err != nil {
		return nil, api.BadRequest
	}

	hourTo, err := getParamInt(ps, "hourTo")
	if err != nil {
		return nil, api.BadRequest
	}

	if monthFrom < 1 || monthFrom > 12 || monthTo < 1 || monthTo > 12 {
		return nil, api.BadRequest
	}

	from := time.Date(yearFrom, time.Month(monthFrom), dayFrom, hourFrom, 0, 0, 0, time.UTC)
	to := time.Date(yearTo, time.Month(monthTo), dayTo, hourTo, 0, 0, 0, time.UTC)

	var response []RangeData
	for t := from; !t.After(to); t = t.Add(time.Hour) {
		data, ok := h.repository.RetrieveHour(t.Year(), t.Month(), t.Day(), t.Hour())
		if !ok {
			return nil, api.InternalServerError
		}
		rangeData := RangeData{
			Time: t,
			Data: data,
		}
		response = append(response, rangeData)

	}
	return response, nil
}

func (h *handler) RangeDaily(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	yearFrom, err := getParamInt(ps, "yearFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	monthFrom, err := getParamInt(ps, "monthFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	dayFrom, err := getParamInt(ps, "dayFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	yearTo, err := getParamInt(ps, "yearTo")
	if err != nil {
		return nil, api.BadRequest
	}

	monthTo, err := getParamInt(ps, "monthTo")
	if err != nil {
		return nil, api.BadRequest
	}

	dayTo, err := getParamInt(ps, "dayTo")
	if err != nil {
		return nil, api.BadRequest
	}

	if monthFrom < 1 || monthFrom > 12 || monthTo < 1 || monthTo > 12 {
		return nil, api.BadRequest
	}

	from := time.Date(yearFrom, time.Month(monthFrom), dayFrom, 0, 0, 0, 0, time.UTC)
	to := time.Date(yearTo, time.Month(monthTo), dayTo, 0, 0, 0, 0, time.UTC)

	var response []RangeData
	for t := from; !t.After(to); t = t.AddDate(0, 0, 1) {
		data, ok := h.repository.RetrieveDay(t.Year(), t.Month(), t.Day())
		if !ok {
			return nil, api.InternalServerError
		}
		rangeData := RangeData{
			Time: t,
			Data: data,
		}
		response = append(response, rangeData)

	}
	return response, nil
}

func (h *handler) RangeMonthly(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	yearFrom, err := getParamInt(ps, "yearFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	monthFrom, err := getParamInt(ps, "monthFrom")
	if err != nil {
		return nil, api.BadRequest
	}

	yearTo, err := getParamInt(ps, "yearTo")
	if err != nil {
		return nil, api.BadRequest
	}

	monthTo, err := getParamInt(ps, "monthTo")
	if err != nil {
		return nil, api.BadRequest
	}

	if monthFrom < 1 || monthFrom > 12 || monthTo < 1 || monthTo > 12 {
		return nil, api.BadRequest
	}

	from := time.Date(yearFrom, time.Month(monthFrom), 0, 0, 0, 0, 0, time.UTC)
	to := time.Date(yearTo, time.Month(monthTo), 0, 0, 0, 0, 0, time.UTC)

	var response []RangeData
	for t := from; !t.After(to); t = t.AddDate(0, 1, 0) {
		data, ok := h.repository.RetrieveMonth(t.Year(), t.Month())
		if !ok {
			return nil, api.InternalServerError
		}
		rangeData := RangeData{
			Time: t,
			Data: data,
		}
		response = append(response, rangeData)

	}
	return response, nil
}

func getParamInt(ps httprouter.Params, name string) (int, error) {
	return strconv.Atoi(getParamString(ps, name))
}

func getParamString(ps httprouter.Params, name string) string {
	return strings.TrimSuffix(ps.ByName(name), ".json")
}

type RangeData struct {
	Time time.Time  `json:"time"`
	Data *core.Data `json:"data"`
}

func Serve(repository *core.Repository, address string) error {
	handler := newHandler(repository)
	// Add CORS middleware.
	handler = cors.AllowAll().Handler(handler)
	// Add GZIP middleware.
	handler = gziphandler.GzipHandler(handler)

	log.Info("starting listening", "address", address)
	return http.ListenAndServe(address, handler)
}

func newHandler(repository *core.Repository) http.Handler {
	h := &handler{
		repository: repository,
	}
	router := httprouter.New()

	// Discrete endpoints
	router.GET("/api/hour/:year/:month/:day/:hour", api.Wrap(h.Hour))
	router.GET("/api/day/:year/:month/:day", api.Wrap(h.Day))
	router.GET("/api/month/:year/:month", api.Wrap(h.Month))

	// Range endpoints
	router.GET("/api/range/hourly/:yearFrom/:monthFrom/:dayFrom/:hourFrom/:yearTo/:monthTo/:dayTo/:hourTo", api.Wrap(h.RangeHourly))
	router.GET("/api/range/daily/:yearFrom/:monthFrom/:dayFrom/:yearTo/:monthTo/:dayTo", api.Wrap(h.RangeDaily))
	router.GET("/api/range/monthly/:yearFrom/:monthFrom/:yearTo/:monthTo", api.Wrap(h.RangeMonthly))

	return router
}
