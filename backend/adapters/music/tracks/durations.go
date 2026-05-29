package tracks

import (
	"context"
	"sync"

	"github.com/boreq/eggplant/application/music"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/errors"
)

type DurationChecker interface {
	GetDuration(ctx context.Context, path string) (musicdomain.TrackDuration, error)
}

type DurationStore struct {
	checker DurationChecker

	mu    sync.Mutex
	paths map[musicdomain.FileId]string                    // known file id -> path, replaced on every reload
	cache map[musicdomain.FileId]musicdomain.TrackDuration // memoised durations
}

func NewDurationStore(checker DurationChecker) *DurationStore {
	return &DurationStore{
		checker: checker,
		paths:   make(map[musicdomain.FileId]string),
		cache:   make(map[musicdomain.FileId]musicdomain.TrackDuration),
	}
}

func (s *DurationStore) SetItems(items []music.TrackStoreItem) {
	paths := make(map[musicdomain.FileId]string, len(items))
	for _, item := range items {
		paths[item.FileId()] = item.Path().String()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.paths = paths
	for fileId := range s.cache {
		if _, ok := paths[fileId]; !ok {
			delete(s.cache, fileId)
		}
	}
}

func (s *DurationStore) GetDuration(ctx context.Context, fileId musicdomain.FileId) (musicdomain.TrackDuration, error) {
	s.mu.Lock()
	if d, ok := s.cache[fileId]; ok {
		s.mu.Unlock()
		return d, nil
	}
	path, ok := s.paths[fileId]
	s.mu.Unlock()

	if !ok {
		return musicdomain.TrackDuration{}, errors.New("unknown file id")
	}

	d, err := s.checker.GetDuration(ctx, path)
	if err != nil {
		return musicdomain.TrackDuration{}, errors.Wrap(err, "could not probe duration")
	}

	s.mu.Lock()
	s.cache[fileId] = d
	s.mu.Unlock()

	return d, nil
}
