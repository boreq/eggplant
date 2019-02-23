package core

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/boreq/goaccess/config"
	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser"
)

const entryKeyFormat = "2006-01-02 15"

func NewRepository(conf *config.Config) *Repository {
	rv := &Repository{
		data: make(map[string]*Data),
		log:  logging.New("repository"),
		conf: conf,
	}
	return rv
}

type Repository struct {
	data      map[string]*Data
	dataMutex sync.Mutex
	conf      *config.Config
	log       logging.Logger
}

func (r *Repository) Insert(entry *parser.Entry) error {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

	r.normalize(entry)

	key := r.createKey(entry.Time)
	data, ok := r.data[key]
	if !ok {
		data = NewData()
		r.data[key] = data
	}
	return data.Insert(entry)
}

func (r *Repository) Retrieve(year int, month time.Month, day int, hour int) (*Data, bool) {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

	t := time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	key := r.createKey(t)
	d, ok := r.data[key]
	return d, ok
}

func (r *Repository) createKey(date time.Time) string {
	return date.UTC().Format(entryKeyFormat)
}

func (r *Repository) normalize(entry *parser.Entry) {
	if r.conf.NormalizeQuery {
		u, err := url.ParseRequestURI(entry.HttpRequestURI)
		if err != nil {
			if entry.Status != "400" {
				r.log.Warn("query normalization failed", "err", err, "entry", entry)
			}
		} else {
			u.RawQuery = ""
			entry.HttpRequestURI = u.String()
		}
	}
	if r.conf.NormalizeSlash {
		if len(entry.HttpRequestURI) > 1 {
			entry.HttpRequestURI = strings.TrimRight(entry.HttpRequestURI, "/")
		}
	}
}
