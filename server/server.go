package server

import (
	"errors"
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

func extractDate(ps httprouter.Params, yearName, monthName, dayName, hourName string) (year, month, day, hour int, err error) {
	year, err = strconv.Atoi(strings.TrimSuffix(ps.ByName(yearName), ".json"))
	if err != nil {
		err = errors.New("invalid year")
		return
	}

	month, err = strconv.Atoi(strings.TrimSuffix(ps.ByName(monthName), ".json"))
	if err != nil {
		err = errors.New("invalid month")
		return
	}

	day, err = strconv.Atoi(strings.TrimSuffix(ps.ByName(dayName), ".json"))
	if err != nil {
		err = errors.New("invalid day")
		return
	}

	hour, err = strconv.Atoi(strings.TrimSuffix(ps.ByName(hourName), ".json"))
	if err != nil {
		err = errors.New("invalid hour")
		return
	}

	if month < 1 || month > 12 {
		err = errors.New("month must be in range [1, 12]")
		return
	}
	return
}

func getParamInt(ps httprouter.Params, name string) (int, error) {
	return strconv.Atoi(getParamString(ps, name))
}

func getParamString(ps httprouter.Params, name string) string {
	return strings.TrimSuffix(ps.ByName(name), ".json")
}

func (h *handler) Hour(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	year, month, day, hour, err := extractDate(ps, "year", "month", "day", "hour")
	if err != nil {
		return nil, api.BadRequest.WithError(err)
	}

	response, ok := h.repository.Retrieve(year, time.Month(month), day, hour)
	if !ok {
		return nil, api.NotFound
	}

	return response, nil
}

func iterDateRange(from, to time.Time) <-chan time.Time {
	c := make(chan time.Time)
	go func() {
		defer close(c)
		for d := truncateToHour(from); !d.After(to); d = d.Add(time.Hour) {
			c <- d
		}
	}()
	return c
}

type RangeData struct {
	Time time.Time `json:"time"`
	Data core.Data `json:"data"`
}

type truncateFn func(t time.Time) time.Time

func truncateToHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func truncateToMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 0, 0, 0, 0, 0, t.Location())
}

func getTruncateFn(groupBy string) (truncateFn, error) {
	switch groupBy {
	case "hourly":
		return truncateToHour, nil
	case "daily":
		return truncateToDay, nil
	case "monthly":
		return truncateToMonth, nil
	default:
		return nil, errors.New("unsupported groupBy parameter")
	}
}

func (h *handler) Range(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	fromTimestamp, err := getParamInt(ps, "from")
	if err != nil {
		return nil, api.BadRequest.WithError(err)
	}

	toTimestamp, err := getParamInt(ps, "to")
	if err != nil {
		return nil, api.BadRequest.WithError(err)
	}

	from := time.Unix(int64(fromTimestamp), 0).UTC()
	to := time.Unix(int64(toTimestamp), 0).UTC()

	truncateFn, err := getTruncateFn(getParamString(ps, "groupBy"))
	if err != nil {
		return nil, api.BadRequest
	}

	var response []RangeData
	for t := range iterDateRange(from, to) {
		data, ok := h.repository.Retrieve(t.Year(), t.Month(), t.Day(), t.Hour())
		if !ok {
			data = core.NewData()
		}
		rangeData := RangeData{Time: truncateFn(t), Data: *data}
		var err error
		response, err = addToResponse(response, rangeData)
		if err != nil {
			log.Error("could not add to response", "err", err)
			return nil, api.InternalServerError
		}
	}

	return response, nil
}

func addToResponse(response []RangeData, rangeData RangeData) ([]RangeData, error) {
	data := findMatchingRangeData(response, rangeData.Time)
	if data == nil {
		response = append(response, RangeData{Time: rangeData.Time})
		data = &response[len(response)-1]
	}
	err := mergeRangeData(&data.Data, rangeData.Data)
	return response, err
}

func findMatchingRangeData(response []RangeData, t time.Time) *RangeData {
	for i := range response {
		if response[i].Time.Equal(t) {
			return &response[i]
		}
	}
	return nil
}

func mergeRangeData(target *core.Data, source core.Data) error {
	// Group referers.
	for _, src := range source.ByReferer {
		targetByReferer := target.GetOrCreateByReferer(src.Referer)
		targetByReferer.InsertHits(src.Hits)
		for _, visit := range src.Visits.Get() {
			targetByReferer.InsertVisit(visit)
		}
	}

	// Group URIs.
	for _, src := range source.ByUri {
		targetByUri := target.GetOrCreateByUri(src.Uri)
		for _, visit := range src.Visits.Get() {
			targetByUri.InsertVisit(visit)
		}
		for _, srcStatus := range src.ByStatus {
			targetByStatus := targetByUri.GetOrCreateByStatus(srcStatus.Status)
			targetByStatus.InsertHits(srcStatus.Hits)
		}
	}

	return nil
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
	router.GET("/api/from/:from/to/:to/:groupBy", api.Wrap(h.Range))
	router.GET("/api/hour/:year/:month/:day/:hour", api.Wrap(h.Hour))
	return router
}
