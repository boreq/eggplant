// Package store is responsible for conversion and storage of tracks and
// thumbnails.
package store

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

// scanEvery specifies the interval in which the store checks if there are any
// items that need to be converted or cleaned up. This is done periodically on
// top of performing this check every time new items are received to make sure
// that the data will be regenerated in case of a failure.
const scanEvery = 30 * time.Minute

// errorDelay specifies the delay after a failed conversion or cleanup. As most
// errors are I/O related this ensures that the store will not be using up too
// much resources attempting to convert the files over and over again and
// encountering the same issue with each conversion.
const errorDelay = 10 * time.Second

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
		itemsCh:   make(chan []Item),
		wg:        &sync.WaitGroup{},
		log:       log,
	}

	ch := make(chan Item)
	s.startWorkers(runtime.NumCPU(), ch)
	go s.run(ch)
	return s, nil
}

type Store struct {
	converter Converter
	itemsCh   chan []Item
	items     []Item
	mutex     sync.Mutex
	wg        *sync.WaitGroup
	log       logging.Logger
}

func (s *Store) GetStats() (queries.StoreStats, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	converted, err := s.countConvertedItems()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not count converted items")
	}

	originalSize, err := s.getOriginalSize()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not get the original size")
	}

	convertedSize, err := s.getConvertedSize()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not get the converted size")
	}

	stats := queries.StoreStats{
		AllItems:       len(s.items),
		ConvertedItems: converted,
		OriginalSize:   originalSize,
		ConvertedSize:  convertedSize,
	}
	return stats, nil
}

func (s *Store) SetItems(items []Item) {
	s.itemsCh <- items
}

func (s *Store) GetFilePath(id string) (string, error) {
	return s.converter.OutputFile(id), nil
}

func (s *Store) startWorkers(n int, ch <-chan Item) {
	s.log.Debug("starting workers", "n", n)
	for i := 0; i < runtime.NumCPU(); i++ {
		go s.worker(ch)
	}
}

func (s *Store) setItems(items []Item) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = items
}

func (s *Store) run(ch chan<- Item) {
	itemsCh := onlyLast(s.itemsCh)
outer:
	for {
		s.log.Debug("starting cleanup and conversion")

		if err := s.performCleanup(); err != nil {
			s.log.Error("could not perform the cleanup", "err", err)
			<-time.After(errorDelay)
			continue
		}

		items := make([]Item, len(s.items))
		copy(items, s.items)

		rand.Shuffle(len(items), func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})

		for _, item := range items {
			convert, err := s.needsConversion(item)
			if err != nil {
				s.log.Error("could not check if an item needs conversion", "err", err)
				<-time.After(errorDelay)
				continue outer
			}

			if convert {
				select {
				case items := <-itemsCh:
					s.wg.Wait()
					s.setItems(items)
					continue outer
				case ch <- item:
					s.wg.Add(1)
				}
			}
		}

		s.wg.Wait()

		select {
		case items := <-itemsCh:
			s.wg.Wait()
			s.setItems(items)
		case <-time.After(scanEvery):
		}
	}
}

func (s *Store) worker(ch <-chan Item) {
	for {
		item, ok := <-ch
		if !ok {
			return
		}

		if err := s.convert(item); err != nil {
			s.log.Error("conversion failed", "err", err)
			<-time.After(errorDelay)
		}
	}
}

func (s *Store) convert(item Item) error {
	defer s.wg.Done()

	if err := s.converter.Convert(item); err != nil {
		return errors.Wrapf(err, "conversion of '%s' failed", item.Path)
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
		if os.IsNotExist(err) {
			return nil
		}
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

func (s *Store) needsConversion(item Item) (bool, error) {
	file := s.converter.OutputFile(item.Id)
	isConverted, err := exists(file)
	if err != nil {
		return false, errors.Wrap(err, "could not check if a file exists")
	}
	return !isConverted, nil
}

func (s *Store) countConvertedItems() (int, error) {
	counter := 0
	for _, item := range s.items {
		needsConversion, err := s.needsConversion(item)
		if err != nil {
			return 0, errors.Wrap(err, "could not check if a file exists")
		}
		if !needsConversion {
			counter++
		}
	}
	return counter, nil
}

func (s *Store) getOriginalSize() (int64, error) {
	var sum int64
	for _, item := range s.items {
		fileInfo, err := os.Stat(item.Path)
		if err != nil {
			return 0, errors.Wrap(err, "could not stat")
		}
		sum += fileInfo.Size()
	}
	return sum, nil
}

func (s *Store) getConvertedSize() (int64, error) {
	var sum int64
	for _, item := range s.items {
		fileInfo, err := os.Stat(s.converter.OutputFile(item.Id))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return 0, errors.Wrap(err, "could not stat")
		}
		sum += fileInfo.Size()
	}
	return sum, nil
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

func onlyLast(inCh <-chan []Item) <-chan []Item {
	ch := make(chan []Item)
	go func() {
		defer close(ch)

		var received *[]Item

		for {
			if received != nil {
				select {
				case item, ok := <-inCh:
					if !ok {
						return
					}
					tmp := item
					received = &tmp
				case ch <- *received:
					received = nil
				}
			} else {
				item, ok := <-inCh
				if !ok {
					return
				}
				tmp := item
				received = &tmp
			}
		}
	}()
	return ch
}
