package core

import (
	"sync"
	"time"

	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser"
)

func NewRepository() *Repository {
	rv := &Repository{
		data: make(map[string]*Data),
		log:  logging.New("repository"),
	}
	return rv
}

type Repository struct {
	data      map[string]*Data
	dataMutex sync.Mutex
	log       logging.Logger
}

func (r *Repository) Insert(entry *parser.Entry) error {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

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

const entryKeyFormat = "2006-01-02 15"

func (r *Repository) createKey(date time.Time) string {
	return date.UTC().Format(entryKeyFormat)
}
