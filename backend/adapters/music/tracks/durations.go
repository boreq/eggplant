package tracks

import (
	"context"
	"runtime"
	"sync"

	"github.com/boreq/eggplant/application/music"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type DurationChecker interface {
	GetDuration(ctx context.Context, path string) (musicdomain.TrackDuration, error)
}

type DurationStore struct {
	ctx     context.Context
	checker DurationChecker
	log     logging.Logger

	mu                sync.Mutex
	paths             map[musicdomain.FileId]string
	cache             map[musicdomain.FileId]musicdomain.TrackDuration
	currentlyChecking map[musicdomain.FileId]*durationResult

	jobs chan durationJob
}

func NewDurationStore(ctx context.Context, checker DurationChecker) *DurationStore {
	s := &DurationStore{
		ctx:               ctx,
		checker:           checker,
		log:               logging.New("tracks.DurationStore"),
		paths:             make(map[musicdomain.FileId]string),
		cache:             make(map[musicdomain.FileId]musicdomain.TrackDuration),
		currentlyChecking: make(map[musicdomain.FileId]*durationResult),
		jobs:              make(chan durationJob),
	}
	for range runtime.NumCPU() {
		go s.worker(ctx)
	}
	return s
}

func (s *DurationStore) SetItems(items []music.TrackStoreItem) {
	paths := make(map[musicdomain.FileId]string, len(items))
	for _, item := range items {
		paths[item.FileId()] = item.Path().String()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.paths = paths
	s.dropCacheForRemovedFiles()
}

func (s *DurationStore) dropCacheForRemovedFiles() {
	for fileId := range s.cache {
		if _, ok := s.paths[fileId]; !ok {
			delete(s.cache, fileId)
		}
	}
}

func (s *DurationStore) GetDuration(ctx context.Context, fileId musicdomain.FileId) (musicdomain.TrackDuration, error) {
	cached, result, err := s.getCachedOrQueueCheck(fileId)
	if err != nil {
		return musicdomain.TrackDuration{}, errors.Wrap(err, "could not get result")
	}
	if result == nil {
		return cached, nil
	}

	select {
	case <-result.done:
		if result.err != nil {
			return musicdomain.TrackDuration{}, errors.Wrap(result.err, "could not probe duration")
		}
		return result.d, nil
	case <-ctx.Done():
		return musicdomain.TrackDuration{}, errors.Wrap(ctx.Err(), "context done while waiting for duration")
	}
}

func (s *DurationStore) getCachedOrQueueCheck(fileId musicdomain.FileId) (musicdomain.TrackDuration, *durationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if d, ok := s.cache[fileId]; ok {
		return d, nil, nil
	}
	if result, ok := s.currentlyChecking[fileId]; ok {
		return musicdomain.TrackDuration{}, result, nil
	}
	path, ok := s.paths[fileId]
	if !ok {
		return musicdomain.TrackDuration{}, nil, errors.New("unknown file id")
	}

	result := &durationResult{done: make(chan struct{})}
	s.currentlyChecking[fileId] = result
	go s.enqueue(durationJob{fileId: fileId, path: path, result: result})
	return musicdomain.TrackDuration{}, result, nil
}

func (s *DurationStore) enqueue(job durationJob) {
	select {
	case s.jobs <- job:
	case <-s.ctx.Done():
		s.fulfil(job.fileId, job.result, musicdomain.TrackDuration{}, s.ctx.Err())
	}
}

func (s *DurationStore) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.jobs:
			d, err := s.checker.GetDuration(ctx, job.path)
			if err != nil && !errors.Is(err, context.Canceled) {
				s.log.Warn("probe failed", "path", job.path, "err", err)
			}
			s.fulfil(job.fileId, job.result, d, err)
		}
	}
}

func (s *DurationStore) fulfil(fileId musicdomain.FileId, result *durationResult, d musicdomain.TrackDuration, err error) {
	s.mu.Lock()
	if err == nil {
		s.cache[fileId] = d
	}
	delete(s.currentlyChecking, fileId)
	s.mu.Unlock()

	result.d = d
	result.err = err
	close(result.done)
}

type durationResult struct {
	done chan struct{}
	d    musicdomain.TrackDuration
	err  error
}

type durationJob struct {
	fileId musicdomain.FileId
	path   string
	result *durationResult
}
