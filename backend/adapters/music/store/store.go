// Package store is responsible for conversion and storage of tracks and
// thumbnails.
package store

import (
	"context"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"

	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

const (
	// cleanupEvery specifies how often the cache directory will be scanned for
	// items that should be removed.
	cleanupEvery = cacheItemsFor / 2

	// cleanupErrorDelay specifies the delay after a failed cleanup. As most
	// errors are I/O related this ensures that the store will not be using up
	// too much resources attempting to convert the files over and over again
	// and encountering the same issue with each conversion.
	cleanupErrorDelay = 1 * time.Minute

	// cacheItemsFor specifies the amount of time that has to pass without the
	// item being accessed for the converted item to be removed.
	cacheItemsFor = 30 * time.Minute
)

type Item struct {
	Id   string
	Path string
}

type Converter interface {
	OutputFile(id string) string
	TemporaryOutputFile(id string) string
	OutputDirectory() string
	Convert(item Item) error
}

type Store struct {
	items            map[string]Item      // key is item id
	itemsAccessTimes map[string]time.Time // key is item id

	ongoingConversions map[string][]scheduledConversion // key is item id
	conversionsCh      chan scheduledConversion

	mutex sync.Mutex

	log       logging.Logger
	converter Converter
}

func NewStore(ctx context.Context, log logging.Logger, converter Converter) (*Store, error) {
	s := &Store{
		items:            make(map[string]Item),
		itemsAccessTimes: make(map[string]time.Time),

		ongoingConversions: make(map[string][]scheduledConversion),
		conversionsCh:      make(chan scheduledConversion),

		log:       log,
		converter: converter,
	}

	s.startConversionWorkers(ctx)
	s.startCleanupWorker(ctx)

	return s, nil
}

type ConvertedFileOrError struct {
	ConvertedFile music.ConvertedFile
	Err           error
}

func (s *Store) GetStats() (queries.StoreStats, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	originalSize, err := s.getOriginalSize()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not get the original size")
	}

	convertedSize, convertedCount, err := s.getConvertedStats()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not get the converted size")
	}

	stats := queries.StoreStats{
		AllItems:       int64(len(s.items)),
		ConvertedItems: convertedCount,
		OriginalSize:   originalSize,
		ConvertedSize:  convertedSize,
	}
	return stats, nil
}

func (s *Store) SetItems(items []Item) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = make(map[string]Item)
	for _, item := range items {
		s.items[item.Id] = item
	}

	for itemId := range s.itemsAccessTimes {
		if _, ok := s.items[itemId]; !ok {
			delete(s.itemsAccessTimes, itemId)
		}
	}
}

func (s *Store) GetConvertedFile(ctx context.Context, id string) (music.ConvertedFile, error) {
	ch := make(chan ConvertedFileOrError, 0)
	go func() {
		convertedFile, err := s.getConvertedFile(ctx, id)
		select {
		case ch <- ConvertedFileOrError{
			ConvertedFile: convertedFile,
			Err:           err,
		}:
		case <-ctx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		return music.ConvertedFile{}, ctx.Err()
	case v := <-ch:
		if err := v.Err; err != nil {
			return music.ConvertedFile{}, errors.Wrap(err, "error getting converted file")
		}
		return v.ConvertedFile, nil
	}
}

func (s *Store) getConvertedFile(ctx context.Context, id string) (music.ConvertedFile, error) {
	f, err := s.getFile(id)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return music.ConvertedFile{}, errors.Wrap(err, "error getting the file")
		}

		errCh := s.scheduleConversion(ctx, id)
		select {
		case err := <-errCh:
			if err != nil {
				return music.ConvertedFile{}, errors.Wrap(err, "conversion error")
			}
		case <-ctx.Done():
			return music.ConvertedFile{}, ctx.Err()
		}

		f, err = s.getFile(id)
		if err != nil {
			return music.ConvertedFile{}, errors.Wrap(err, "error getting the file again")
		}
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return music.ConvertedFile{}, errors.Wrap(err, "stat error")
	}

	return music.ConvertedFile{
		Name:    f.Name(),
		Modtime: fileInfo.ModTime(),
		Content: f,
	}, nil
}

func (s *Store) getFile(id string) (*os.File, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.items[id]; !ok {
		return nil, errors.New("item does not exist")
	}

	s.itemsAccessTimes[id] = time.Now()
	return os.Open(s.converter.OutputFile(id))
}

func (s *Store) scheduleConversion(ctx context.Context, id string) <-chan error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	errCh := make(chan error)
	conversion := scheduledConversion{
		ItemId: id,
		Ctx:    ctx,
		ErrCh:  errCh,
	}

	go func() {
		select {
		case s.conversionsCh <- conversion:
		case <-ctx.Done():
			return
		}
	}()

	return errCh
}

func (s *Store) startConversionWorkers(ctx context.Context) {
	for i := 0; i < runtime.NumCPU(); i++ {
		go s.conversionWorker(ctx)
	}
}

func (s *Store) startCleanupWorker(ctx context.Context) {
	go s.cleanupWorker(ctx)
}

func (s *Store) conversionWorker(ctx context.Context) {
	for {
		select {
		case conversion := <-s.conversionsCh:
			if err := s.convert(conversion); err != nil {
				s.log.Error("conversion failed", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Store) cleanupWorker(ctx context.Context) {
	for {
		if err := s.cleanup(); err != nil {
			s.log.Error("cleanup failed", "err", err)
			select {
			case <-time.After(cleanupErrorDelay):
				continue
			case <-ctx.Done():
				return
			}
		}

		select {
		case <-time.After(cleanupEvery):
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (s *Store) convert(conversion scheduledConversion) (err error) {
	if shouldConvert := s.beginConversion(conversion); !shouldConvert {
		return nil
	}
	defer s.endConversion(conversion, err)

	item, ok := s.items[conversion.ItemId]
	if !ok {
		return errors.New("item does not exist")
	}

	exists, err := s.outputFileExists(item)
	if err != nil {
		return errors.Wrap(err, "error checking if file exists")
	}

	if exists {
		return nil
	}

	start := time.Now()
	defer func() {
		s.log.Debug("conversion ended", "err", err, "duration", time.Since(start))
	}()

	if err := s.converter.Convert(item); err != nil {
		return errors.Wrapf(err, "conversion of '%s' failed", item.Path)
	}

	return nil
}

func (s *Store) beginConversion(item scheduledConversion) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, isAlreadyBeingConverted := s.ongoingConversions[item.ItemId]; isAlreadyBeingConverted {
		s.ongoingConversions[item.ItemId] = append(s.ongoingConversions[item.ItemId], item)
		return false
	}

	s.ongoingConversions[item.ItemId] = []scheduledConversion{item}
	return true
}

func (s *Store) endConversion(conversion scheduledConversion, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, scheduledConversion := range s.ongoingConversions[conversion.ItemId] {
		select {
		case scheduledConversion.ErrCh <- err:
		case <-scheduledConversion.Ctx.Done():
		}
	}

	delete(s.ongoingConversions, conversion.ItemId)
}

func (s *Store) cleanup() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Debug("performing cleanup", "dir", s.converter.OutputDirectory())

	filesThatCanExist := make(map[string]struct{})

	for itemId, accessTime := range s.itemsAccessTimes {
		if time.Since(accessTime) < cacheItemsFor {
			filesThatCanExist[s.converter.OutputFile(itemId)] = struct{}{}
		}
	}

	for itemId := range s.ongoingConversions {
		filesThatCanExist[s.converter.OutputFile(itemId)] = struct{}{}
		filesThatCanExist[s.converter.TemporaryOutputFile(itemId)] = struct{}{}
	}

	dir := s.converter.OutputDirectory()
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "could not read the output directory")
	}

	for _, fileInfo := range fileInfos {
		file := path.Join(dir, fileInfo.Name())
		if _, canExist := filesThatCanExist[file]; !canExist {
			s.log.Debug("removing a file", "file", file)
			if err := os.RemoveAll(file); err != nil {
				return errors.Wrap(err, "remove all error")
			}
		}
	}

	return nil
}

func (s *Store) outputFileExists(item Item) (bool, error) {
	file := s.converter.OutputFile(item.Id)
	return exists(file)
}

func (s *Store) ensureOutputDirectoryExists() error {
	dir := s.converter.OutputDirectory()
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.Mkdir(dir, os.ModePerm)
		}
	}
	return nil
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

func (s *Store) getConvertedStats() (size int64, count int64, err error) {
	dirEntries, err := os.ReadDir(s.converter.OutputDirectory())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, 0, nil
		}
		return 0, 0, errors.Wrap(err, "error reading directory")
	}

	for _, dirEntry := range dirEntries {
		fileInfo, err := dirEntry.Info()
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return 0, 0, errors.Wrap(err, "could not stat")
		}
		size += fileInfo.Size()
		count += 1
	}
	return size, count, nil
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

type scheduledConversion struct {
	ItemId string
	Ctx    context.Context
	ErrCh  chan<- error
}
