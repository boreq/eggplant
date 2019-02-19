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

func iterDateRange(yearA, monthA, dayA, hourA, yearB, monthB, dayB, hourB int) <-chan []int {
	dateA := time.Date(yearA, time.Month(monthA), dayA, hourA, 0, 0, 0, time.UTC)
	dateB := time.Date(yearB, time.Month(monthB), dayB, hourB, 0, 0, 0, time.UTC)

	c := make(chan []int)
	go func() {
		defer close(c)
		for d := dateA; !d.After(dateB); d = d.Add(time.Hour) {
			c <- []int{d.Year(), int(d.Month()), d.Day(), d.Hour()}
		}
	}()
	return c
}

type RangeData struct {
	Time TimeData  `json:"time"`
	Data core.Data `json:"data"`
}

type TimeData struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
	Hour  int `json:"hour"`
}

func (h *handler) Range(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	yearA, monthA, dayA, hourA, err := extractDate(ps, "yearA", "monthA", "dayA", "hourA")
	if err != nil {
		return nil, api.BadRequest.WithError(err)
	}

	yearB, monthB, dayB, hourB, err := extractDate(ps, "yearB", "monthB", "dayB", "hourB")
	if err != nil {
		return nil, api.BadRequest.WithError(err)
	}

	var response []RangeData
	for slice := range iterDateRange(yearA, monthA, dayA, hourA, yearB, monthB, dayB, hourB) {
		log.Info("iterating over days", "slice", slice)
		year := slice[0]
		month := slice[1]
		day := slice[2]
		hour := slice[3]
		rd := RangeData{
			Time: TimeData{
				Year:  year,
				Month: month,
				Day:   day,
				Hour:  hour,
			},
		}
		data, ok := h.repository.Retrieve(year, time.Month(month), day, hour)
		if !ok {
			data = core.NewData()
		}
		rd.Data = *data
		response = append(response, rd)
	}

	return response, nil
}

func Serve(repository *core.Repository, address string) error {
	h := &handler{
		repository: repository,
	}
	router := httprouter.New()
	router.GET("/api/hour/:year/:month/:day/:hour", api.Wrap(h.Hour))
	router.GET("/api/range/:yearA/:monthA/:dayA/:hourA/:yearB/:monthB/:dayB/:hourB", api.Wrap(h.Range))
	log.Info("starting listening", "address", address)
	handler := cors.AllowAll().Handler(router)
	handler = gziphandler.GzipHandler(handler)
	return http.ListenAndServe(address, handler)
}
