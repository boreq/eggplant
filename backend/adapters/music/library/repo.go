package library

import (
	"sync/atomic"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain/library"
)

type InMemoryRepository struct {
	current atomic.Pointer[library.Library]
}

func NewInMemoryRepository() *InMemoryRepository {
	r := &InMemoryRepository{}
	return r
}

func (r *InMemoryRepository) Get() (*library.Library, error) {
	lib := r.current.Load()
	if lib == nil {
		return nil, music.ErrLibraryNotReady
	}
	return lib, nil
}

func (r *InMemoryRepository) Save(library *library.Library) {
	r.current.Store(library)
}
