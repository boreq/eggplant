package core

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/boreq/plum/config"
	"github.com/boreq/plum/logging"
	"github.com/boreq/plum/parser"
)

const entryKeyFormat = "2006-01-02 15"

// visitPrefixFormat is used to generate a visit prefix which prevents
// identical visits from different days from getting merged.
const visitPrefixFormat = "2006-01-02"

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

func (r *Repository) RetrieveHour(year int, month time.Month, day int, hour int) (*Data, bool) {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

	target := NewData()

	t := time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
	key := r.createKey(t)
	if d, ok := r.data[key]; ok {
		visitPrefix := t.Format(visitPrefixFormat)
		mergeData(target, d, visitPrefix)
	}
	return target, true
}

func (r *Repository) RetrieveDay(year int, month time.Month, day int) (*Data, bool) {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

	target := NewData()

	for _, t := range iterateDay(year, month, day) {
		key := r.createKey(t)
		if d, ok := r.data[key]; ok {
			visitPrefix := t.Format(visitPrefixFormat)
			mergeData(target, d, visitPrefix)
		}
	}
	return target, true
}

func (r *Repository) RetrieveMonth(year int, month time.Month) (*Data, bool) {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

	target := NewData()

	for _, t := range iterateMonth(year, month) {
		key := r.createKey(t)
		if d, ok := r.data[key]; ok {
			visitPrefix := t.Format(visitPrefixFormat)
			mergeData(target, d, visitPrefix)
		}
	}
	return target, true
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
	if r.conf.StripRefererProtocol {
		entry.Referer = strings.TrimPrefix(entry.Referer, "http://")
		entry.Referer = strings.TrimPrefix(entry.Referer, "https://")
	}
}

func iterateDay(year int, month time.Month, day int) []time.Time {
	var result []time.Time
	start := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month, day+1, 0, 0, 0, 0, time.UTC)
	for t := start; t.Before(end); t = t.Add(time.Hour) {
		result = append(result, t)
	}
	return result
}

func iterateMonth(year int, month time.Month) []time.Time {
	var result []time.Time
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
	for t := start; t.Before(end); t = t.Add(time.Hour) {
		result = append(result, t)
	}
	return result
}

func mergeData(target *Data, source *Data, visitPrefix string) {
	// Copy visits
	for visit := range source.Visits {
		target.InsertVisit(visitPrefix + visit)
	}

	// Group referers
	for sourceReferer, sourceRefererData := range source.Referers {
		targetRefererData := target.GetOrCreateRefererData(sourceReferer)
		targetRefererData.InsertHits(sourceRefererData.Hits)
		for visit := range sourceRefererData.Visits {
			targetRefererData.InsertVisit(visitPrefix + visit)
		}
	}

	// Group URIs
	for sourceUri, sourceUriData := range source.Uris {
		targetUriData := target.GetOrCreateUriData(sourceUri)
		for visit := range sourceUriData.Visits {
			targetUriData.InsertVisit(visitPrefix + visit)
		}
		targetUriData.AddBodyBytesSent(sourceUriData.BodyBytesSent)
		for sourceStatus, sourceStatusData := range sourceUriData.Statuses {
			targetStatusData := targetUriData.GetOrCreateStatusData(sourceStatus)
			targetStatusData.InsertHits(sourceStatusData.Hits)
		}
	}

}
