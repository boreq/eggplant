package core

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser"
	"github.com/dustin/go-humanize"
)

func NewRepository() *Repository {
	rv := &Repository{
		data: make(map[string]*Data),
		log:  logging.New("repository"),
	}
	go func() {
		for range time.Tick(1 * time.Second) {
			rv.printStats()
		}
	}()
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

func (r *Repository) printStats() {
	r.dataMutex.Lock()
	defer r.dataMutex.Unlock()

	var hashSize = visitHash.New().Size()

	m := &memoryTracker{}

	for key, data := range r.data {
		m.record("key", len(key))
		m.record("data", int(reflect.TypeOf(*data).Size()))

		for refKey, refData := range data.ByReferer {
			m.record("refKey", len(refKey))
			m.record("refData", int(reflect.TypeOf(*refData).Size()))
			//r.log.Debug("data", "hashsize", hashSize, "visits", refData.Visits.Size())
			for i := 0; i < refData.Visits.Size(); i++ {
				m.record("refData.Visits", hashSize)
			}
			m.record("refData.Hits", int(reflect.TypeOf(refData.Hits).Size()))
		}

		for uriKey, uriData := range data.ByUri {
			m.record("uriKey", len(uriKey))
			m.record("uriData", int(reflect.TypeOf(*uriData).Size()))
			//m.record("uriData.Visits", uriData.Visits.Size()*hashSize)
			for i := 0; i < uriData.Visits.Size(); i++ {
				m.record("uriData.Visits", hashSize)
			}

			for statKey, statData := range uriData.ByStatus {
				m.record("uriData.ByStatus.statKey", len(statKey))
				m.record("uriData.ByStatus.statData", int(reflect.TypeOf(statData).Size()))
			}
		}
	}

	runtime.GC()
	m.display()
	printMemUsage()
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %s", humanize.Bytes(m.Alloc))
	fmt.Printf("\tTotalAlloc = %s", humanize.Bytes(m.TotalAlloc))
	fmt.Printf("\tSys = %s", humanize.Bytes(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

type memoryData struct {
	Name   string
	Amount int
	Size   int
}

type memoryTracker struct {
	data []*memoryData
}

func (m *memoryTracker) record(name string, size int) {
	data := m.findOrCreate(name)
	data.Amount += 1
	data.Size += size
}

func (m *memoryTracker) findOrCreate(name string) *memoryData {
	for _, data := range m.data {
		if data.Name == name {
			return data
		}
	}
	data := &memoryData{
		Name: name,
	}
	m.data = append(m.data, data)
	return data
}

func (m *memoryTracker) display() {
	total := 0
	for _, data := range m.data {
		total += data.Size
	}

	for _, data := range m.data {
		fmt.Printf("%40s: %10d (%7s - %7.2f%% - %7.2f bytes per item)\n", data.Name, data.Amount, humanize.Bytes(uint64(data.Size)), float32(data.Size)/float32(total)*100, float32(data.Size)/float32(data.Amount))
	}

	fmt.Printf("#total: %s\n", humanize.Bytes(uint64(total)))
}
