package store

import (
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
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

func NewStore(converter Converter) (*Store, error) {
	s := &Store{
		converter: converter,
		ch:        make(chan []Item),
		log:       logging.New("store"),
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

func (s *Store) SetItems(items []Item) {
	s.ch <- items
}

func (s *Store) ServeFile(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, s.converter.OutputFile(id))
}

func (s *Store) receive() {
	for items := range s.ch {
		if err := s.handle(items); err != nil {
			s.log.Error("could not handle updates", "err", err)
		}
	}
}

func (s *Store) handle(items []Item) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = items
	s.itemsSet = true
	return nil
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
			s.log.Error("conversion failed", "err", err)
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

const scanEvery = 10 * time.Second

func makeDirectory(file string) error {
	dir, _ := path.Split(file)
	return os.MkdirAll(dir, os.ModePerm)
}
