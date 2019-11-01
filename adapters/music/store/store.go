// Package store is responsible for conversion and storage of tracks and
// thumbnails.
package store

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

// scanEvery specifies the interval in which the store checks if there are any
// items that need to be converted. This is done periodically to make sure that
// the data will be regenerated in case of some kind of a failure.
const scanEvery = 60 * time.Second

// errorDelay specifies the delay after a failed conversion or cleanup. As most
// errors are I/O related this ensures that the store will not be using up too
// much resources attempting to convert the files over and over again and
// encountering the same issue with each conversion.
const errorDelay = 10 * time.Second

// cleanupDelay specifies how much time has to pass without receiving any items
// for the store to start a cleanup process. This is done to make sure that the
// cleanups don't occur too often.
const cleanupDelay = 60 * time.Second

type Stats struct {
	AllItems       int `json:"allItems"`
	ConvertedItems int `json:"convertedItems"`
}

type Item struct {
	Id   string
	Path string
}

type Converter interface {
	OutputFile(id string) string
	OutputDirectory() string
	Convert(item Item) error
}

func NewStore(log logging.Logger, converter Converter) (*Store, error) {
	s := &Store{
		converter: converter,
		ch:        make(chan []Item),
		log:       log,
	}
	go s.receive()
	go s.process()
	return s, nil
}

type Store struct {
	converter   Converter
	ch          chan []Item
	items       []Item
	nextCleanup time.Time
	mutex       sync.Mutex
	log         logging.Logger
}

func (s *Store) GetStats() (Stats, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	converted, err := s.countConvertedItems()
	if err != nil {
		return Stats{}, errors.Wrap(err, "could not count converted items")
	}

	stats := Stats{
		AllItems:       len(s.items),
		ConvertedItems: converted,
	}
	return stats, nil
}

func (s *Store) SetItems(items []Item) {
	s.ch <- items
}

func (s *Store) ServeFile(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, s.converter.OutputFile(id))
}

func (s *Store) receive() {
	for items := range s.ch {
		s.handleReceivedItems(items)
	}
}

func (s *Store) handleReceivedItems(items []Item) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Debug("received items")
	s.items = items
	s.scheduleCleanup()
}

func (s *Store) scheduleCleanup() {
	s.nextCleanup = time.Now().Add(cleanupDelay)
}

func (s *Store) process() {
	for {
		if err := s.considerConversion(); err != nil {
			s.log.Error("conversion failed", "err", err)
			<-time.After(errorDelay)
		}

		if err := s.considerCleanup(); err != nil {
			s.log.Error("cleanup failed", "err", err)
			<-time.After(errorDelay)
		}
	}
}

func (s *Store) considerConversion() error {
	item, ok, err := s.getNextItem()
	if err != nil {
		return errors.Wrap(err, "could not get a next item for conversion")
	}

	if !ok {
		s.log.Debug("no items to convert")
		<-time.After(scanEvery)
		return nil
	}

	s.log.Debug("converting an item", "item", item)
	if err := s.converter.Convert(item); err != nil {
		return errors.Wrapf(err, "conversion of '%s' failed", item.Path)
	}

	return nil
}

func (s *Store) considerCleanup() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Debug("considering a cleanup")
	if !s.nextCleanup.IsZero() && time.Now().After(s.nextCleanup) {
		s.nextCleanup = time.Time{}
		s.log.Debug("performing a cleanup")
		if err := s.performCleanup(); err != nil {
			return errors.Wrap(err, "cleanup error")
		}
	}
	return nil
}

func (s *Store) performCleanup() error {
	items := make(map[string]bool)

	for _, item := range s.items {
		file := s.converter.OutputFile(item.Id)
		items[file] = true
	}

	dir := s.converter.OutputDirectory()
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.Wrap(err, "could not read the output directory")
	}

	for _, fileInfo := range fileInfos {
		file := path.Join(dir, fileInfo.Name())
		if _, shouldExist := items[file]; !shouldExist {
			s.log.Debug("removing a file", "file", file)
			if err := os.RemoveAll(file); err != nil {
				return errors.Wrap(err, "remove all error")
			}
		}
	}

	return nil
}

func (s *Store) getNextItem() (Item, bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rand.Shuffle(len(s.items), func(i, j int) {
		s.items[i], s.items[j] = s.items[j], s.items[i]
	})

	for _, item := range s.items {
		file := s.converter.OutputFile(item.Id)
		isConverted, err := exists(file)
		if err != nil {
			return Item{}, false, errors.Wrap(err, "could not check if a file exists")
		}
		if !isConverted {
			return item, true, nil
		}
	}
	return Item{}, false, nil
}

func (s *Store) countConvertedItems() (int, error) {
	counter := 0
	for _, item := range s.items {
		file := s.converter.OutputFile(item.Id)
		isConverted, err := exists(file)
		if err != nil {
			return 0, errors.Wrap(err, "could not check if a file exists")
		}
		if isConverted {
			counter++
		}
	}
	return counter, nil
}

func exists(file string) (bool, error) {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "could not stat")
	}
	return true, nil
}

func makeDirectory(file string) error {
	dir, _ := path.Split(file)
	return os.MkdirAll(dir, os.ModePerm)
}
