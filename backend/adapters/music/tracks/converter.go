package tracks

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

const (
	playlistFilename = "playlist.m3u8"
	initFilename     = "init"
	segmentTemplate  = "%d"
	streamsDirectory = "streams"

	hlsSegmentSeconds    = 4
	readinessTimeout     = 30 * time.Second
	readinessPoll        = 50 * time.Millisecond
	readinessFragments   = 5
	jobAcceptanceTimeout = 30 * time.Second

	streamIdleTimeout = 5 * time.Minute
	cleanupInterval   = 1 * time.Minute

	maxOpenStreams = 1000
)

var readinessLastFragmentId = domain.MustNewFragmentId(readinessFragments - 1)

type Converter struct {
	ctx     context.Context
	dataDir string
	log     logging.Logger

	mu      sync.Mutex
	items   map[domain.FileId]music.TrackStoreItem
	streams map[string]*stream

	ffmpegJobs chan ffmpegJob
}

func NewConverter(ctx context.Context, dataDir string) (*Converter, error) {
	c := &Converter{
		ctx:        ctx,
		dataDir:    dataDir,
		log:        logging.New("tracks.Converter"),
		items:      make(map[domain.FileId]music.TrackStoreItem),
		streams:    make(map[string]*stream),
		ffmpegJobs: make(chan ffmpegJob),
	}
	if err := c.removeLeftoverStreamDirectories(); err != nil {
		return nil, err
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go c.ffmpegWorker(ctx)
	}
	go c.cleanupIdleStreamsLoop(ctx)
	return c, nil
}

func (c *Converter) cleanupIdleStreamsLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(cleanupInterval):
			c.cancelIdleStreams()
		}
	}
}

func (c *Converter) cancelIdleStreams() {
	for _, id := range c.unregisterIdleStreams() {
		c.removeStreamDirectory(id)
	}
}

func (c *Converter) unregisterIdleStreams() []domain.StreamId {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	var ids []domain.StreamId
	for key, stream := range c.streams {
		if now.Sub(stream.lastAccess) <= streamIdleTimeout {
			continue
		}
		stream.log.Debug("cleaning up idle stream")
		stream.cancel()
		ids = append(ids, stream.id)
		delete(c.streams, key)
	}
	return ids
}

func (c *Converter) ffmpegWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-c.ffmpegJobs:
			start := time.Now()
			err := c.runFFmpeg(job.s.ctx, job.s)
			job.s.log.Debug("ffmpeg exited", "duration", time.Since(start), "err", err)

			select {
			case job.result <- err:
			case <-job.s.ctx.Done():
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *Converter) SetItems(items []music.TrackStoreItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[domain.FileId]music.TrackStoreItem)
	for _, it := range items {
		c.items[it.FileId()] = it
	}
}

func (c *Converter) StartStream(reqCtx context.Context, fileId domain.FileId, seekPos *domain.SeekPosition) (domain.StreamId, error) {
	stream, err := c.createAndRegisterStream(fileId, seekPos)
	if err != nil {
		return domain.StreamId{}, errors.Wrap(err, "could not register a stream")
	}

	go c.runConversion(stream)

	select {
	case err := <-stream.ready:
		if err != nil {
			return domain.StreamId{}, errors.Wrap(err, "received an error")
		}
		return stream.id, nil
	case <-reqCtx.Done():
		stream.cancel()
		return domain.StreamId{}, reqCtx.Err()
	}
}

func (c *Converter) createAndRegisterStream(fileId domain.FileId, seekPos *domain.SeekPosition) (*stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.streams) >= maxOpenStreams {
		return nil, music.ErrTooManyOpenStreams
	}

	it, ok := c.items[fileId]
	if !ok {
		return nil, errors.New("item does not exist")
	}

	streamId, err := domain.NewStreamId()
	if err != nil {
		return nil, errors.Wrap(err, "could not create stream id")
	}

	ctx, cancel := context.WithCancel(c.ctx)
	s := &stream{
		id:      streamId,
		item:    it,
		seekPos: seekPos,
		ready:   make(chan error),
		log:     c.log.New("trackId", fileId.String(), "streamId", streamId.String()),
		ctx:     ctx,
		cancel:  cancel,
	}
	s.lastAccess = time.Now()
	c.streams[streamId.String()] = s
	return s, nil
}

func (c *Converter) GetPlaylist(fileId domain.FileId, streamId domain.StreamId) (music.ConvertedFile, error) {
	s, err := c.getStreamForFile(streamId, fileId)
	if err != nil {
		return music.ConvertedFile{}, err
	}
	return openFile(c.playlistPathInDir(c.streamDir(s.id)))
}

func (c *Converter) GetInit(fileId domain.FileId, streamId domain.StreamId) (music.ConvertedFile, error) {
	s, err := c.getStreamForFile(streamId, fileId)
	if err != nil {
		return music.ConvertedFile{}, err
	}
	return openFile(c.initPathInDir(c.streamDir(s.id)))
}

func (c *Converter) GetFragment(fileId domain.FileId, streamId domain.StreamId, fragmentId domain.FragmentId) (music.ConvertedFile, error) {
	s, err := c.getStreamForFile(streamId, fileId)
	if err != nil {
		return music.ConvertedFile{}, err
	}
	return openFile(c.fragmentPathInDir(c.streamDir(s.id), fragmentId))
}

func (c *Converter) GetStats() (queries.TrackStats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var originalSize int64
	for _, it := range c.items {
		info, err := os.Stat(it.Path().String())
		if err != nil {
			return queries.TrackStats{}, errors.Wrap(err, "could not stat original")
		}
		originalSize += info.Size()
	}

	convertedSize, convertedCount, err := c.streamsStats()
	if err != nil {
		return queries.TrackStats{}, errors.Wrap(err, "could not read stream stats")
	}

	return queries.TrackStats{
		NumberOfTracks: int64(len(c.items)),
		SizeOfTracks:   originalSize,

		NumberOfStreams:       convertedCount,
		SizeOfConvertedTracks: convertedSize,
	}, nil
}

func (c *Converter) streamsStats() (size int64, count int64, err error) {
	entries, err := os.ReadDir(c.streamsRoot())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		count++
		subEntries, err := os.ReadDir(path.Join(c.streamsRoot(), entry.Name()))
		if err != nil {
			continue
		}
		for _, sub := range subEntries {
			info, err := sub.Info()
			if err != nil {
				continue
			}
			size += info.Size()
		}
	}
	return size, count, nil
}

func (c *Converter) getStreamForFile(streamId domain.StreamId, fileId domain.FileId) (*stream, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.streams[streamId.String()]
	if !ok {
		return nil, errors.New("stream does not exist")
	}
	if s.item.FileId() != fileId {
		return nil, errors.New("stream does not belong to this track")
	}
	s.updateLastAccess()
	return s, nil
}

func (c *Converter) runConversion(s *stream) {
	defer s.cancel()

	ctx := s.ctx

	if err := os.MkdirAll(c.streamDir(s.id), 0755); err != nil {
		s.signalReady(ctx, errors.Wrap(err, "could not create the stream directory"))
		return
	}

	ffmpegResult := make(chan error, 1)
	select {
	case c.ffmpegJobs <- ffmpegJob{s: s, result: ffmpegResult}:
	case <-time.After(jobAcceptanceTimeout):
		s.signalReady(ctx, errors.New("timeout waiting for the worker to pick up a job"))
		return
	case <-ctx.Done():
		return
	}

	readyErr := c.waitForReadiness(ctx, s, ffmpegResult)
	s.signalReady(ctx, readyErr)
	if readyErr != nil {
		if !errors.Is(readyErr, context.Canceled) {
			s.log.Error("readiness failed", "err", readyErr)
		}
		return
	}

	select {
	case err := <-ffmpegResult:
		if err != nil && !errors.Is(err, context.Canceled) {
			s.log.Error("ffmpeg failed", "err", err)
			return
		}
	case <-ctx.Done():
		return
	}
}

func (c *Converter) waitForReadiness(ctx context.Context, s *stream, ffmpegResult <-chan error) error {
	dir := c.streamDir(s.id)
	playlist := c.playlistPathInDir(dir)
	initSegment := c.initPathInDir(dir)
	lastSegment := c.fragmentPathInDir(dir, readinessLastFragmentId)

	tick := time.NewTicker(readinessPoll)
	defer tick.Stop()
	timeout := time.After(readinessTimeout)

	for {
		select {
		case err := <-ffmpegResult:
			if err != nil {
				return errors.Wrap(err, "ffmpeg failed")
			}
			return nil
		case <-timeout:
			return errors.New("timed out waiting for ffmpeg to produce initial output")
		case <-tick.C:
			if fileExists(playlist) && fileExists(initSegment) && fileExists(lastSegment) {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Converter) removeStreamDirectory(id domain.StreamId) {
	dir := c.streamDir(id)
	if err := os.RemoveAll(dir); err != nil {
		c.log.Error("could not remove stream dir", "err", err, "dir", dir)
	}
}

func (s *stream) signalReady(ctx context.Context, err error) {
	select {
	case s.ready <- err:
	case <-ctx.Done():
	}
}

func (c *Converter) runFFmpeg(ctx context.Context, s *stream) error {
	args := []string{"-y"}
	if s.seekPos != nil {
		args = append(args, "-ss", fmt.Sprintf("%f", s.seekPos.Duration().Seconds()))
	}
	args = append(args,
		"-i", s.item.Path().String(),
		"-vn",
		"-c:a", "libopus",
		"-b:a", "96K",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", hlsSegmentSeconds),
		"-hls_playlist_type", "event",
		"-hls_segment_type", "fmp4",
		"-hls_fmp4_init_filename", initFilename,
		"-hls_segment_filename", path.Join(c.streamDir(s.id), segmentTemplate),
		"-hls_base_url", "fragment/",
		c.playlistPathInDir(c.streamDir(s.id)),
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	s.log.Debug("converting", "command", cmd.String())

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "ffmpeg failed (stderr: %s)", strings.TrimSpace(stderr.String()))
	}
	return nil
}

func (c *Converter) streamsRoot() string {
	return path.Join(c.dataDir, streamsDirectory)
}

func (c *Converter) streamDir(id domain.StreamId) string {
	return path.Join(c.streamsRoot(), id.String())
}

func (c *Converter) playlistPathInDir(dir string) string {
	return path.Join(dir, playlistFilename)
}

func (c *Converter) initPathInDir(dir string) string {
	return path.Join(dir, initFilename)
}

func (c *Converter) fragmentPathInDir(dir string, n domain.FragmentId) string {
	return path.Join(dir, fmt.Sprintf(segmentTemplate, n.Int()))
}

func (c *Converter) removeLeftoverStreamDirectories() error {
	if err := os.RemoveAll(c.streamsRoot()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.Wrap(err, "could not clear streams root")
	}
	return nil
}

func openFile(p string) (music.ConvertedFile, error) {
	f, err := os.Open(p)
	if err != nil {
		return music.ConvertedFile{}, errors.Wrapf(err, "could not open '%s'", p)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return music.ConvertedFile{}, errors.Wrap(err, "stat failed")
	}
	return music.ConvertedFile{
		Name:    f.Name(),
		Modtime: info.ModTime(),
		Content: f,
	}, nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

type stream struct {
	id      domain.StreamId
	item    music.TrackStoreItem
	seekPos *domain.SeekPosition
	log     logging.Logger

	ctx        context.Context
	cancel     context.CancelFunc
	lastAccess time.Time

	ready chan error
}

func (s *stream) updateLastAccess() {
	s.lastAccess = time.Now()
}

type ffmpegJob struct {
	s      *stream
	result chan<- error
}
