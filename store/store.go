package store

import (
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/pkg/errors"
)

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
	converter Converter
	ch        chan []Item
	items     []Item
	itemsSet  bool
	mutex     sync.Mutex
	log       logging.Logger
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

	s.items = items
	s.itemsSet = true
}

func (s *Store) process() {
	for {
		item, ok := s.getNextItem()
		if !ok {
			s.log.Debug("no items to convert")
			<-time.After(scanEvery)
			continue
		}

		s.log.Debug("converting an item", "item", item)
		if err := s.converter.Convert(item); err != nil {
			s.log.Error("conversion failed", "err", err, "item", item)
			<-time.After(time.Second)
		}
	}
}

func (s *Store) getNextItem() (Item, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rand.Shuffle(len(s.items), func(i, j int) {
		s.items[i], s.items[j] = s.items[j], s.items[i]
	})

	for _, item := range s.items {
		p := s.converter.OutputFile(item.Id)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return item, true
		}
	}
	return Item{}, false
}

func (s *Store) countConvertedItems() (int, error) {
	counter := 0
	for _, item := range s.items {
		p := s.converter.OutputFile(item.Id)
		if _, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				return 0, err
			}
			continue // isn't converted
		}
		counter++
	}
	return counter, nil
}

const scanEvery = 30 * time.Second

func makeDirectory(file string) error {
	dir, _ := path.Split(file)
	return os.MkdirAll(dir, os.ModePerm)
}
