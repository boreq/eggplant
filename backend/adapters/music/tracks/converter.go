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

	maxOpenStreams = 500
)

var readinessLastFragmentId = domain.MustNewFragmentId(readinessFragments - 1)

type Converter struct {
	dataDir string
	log     logging.Logger

	mu      sync.Mutex
	items   map[domain.FileId]music.TrackStoreItem
	streams map[string]*streamSession

	ffmpegJobs chan ffmpegJob
}

func NewConverter(ctx context.Context, dataDir string) (*Converter, error) {
	c := &Converter{
		dataDir:    dataDir,
		log:        logging.New("tracks.Converter"),
		items:      make(map[domain.FileId]music.TrackStoreItem),
		streams:    make(map[string]*streamSession),
		ffmpegJobs: make(chan ffmpegJob),
	}
	if err := c.removeLeftoverStreamDirectories(); err != nil {
		return nil, err
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go c.ffmpegWorker(ctx)
	}
	return c, nil
}

func (c *Converter) ffmpegWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-c.ffmpegJobs:
			ffmpegCtx, cancel := mergeContexts(job.ctx, ctx)

			start := time.Now()
			err := c.runFFmpeg(ffmpegCtx, job.s)
			cancel()
			job.s.log.Debug("ffmpeg exited", "duration", time.Since(start), "err", err)

			select {
			case job.result <- err:
			case <-job.ctx.Done():
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

func (c *Converter) StartStream(ctx context.Context, fileId domain.FileId, seekPos *domain.SeekPosition) (domain.StreamId, error) {
	streamSession, err := c.createAndRegisterStream(fileId, seekPos)
	if err != nil {
		return domain.StreamId{}, errors.Wrap(err, "could not register a stream")
	}

	go c.run(ctx, streamSession)

	select {
	case err := <-streamSession.ready:
		if err != nil {
			return domain.StreamId{}, errors.Wrap(err, "received an error")
		}
		return streamSession.id, nil
	case <-ctx.Done():
		return domain.StreamId{}, ctx.Err()
	}
}

func (c *Converter) createAndRegisterStream(fileId domain.FileId, seekPos *domain.SeekPosition) (*streamSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.streams) >= maxOpenStreams {
		return nil, errors.New("too many open streams")
	}

	it, ok := c.items[fileId]
	if !ok {
		return nil, errors.New("item does not exist")
	}

	streamId, err := domain.NewStreamId()
	if err != nil {
		return nil, errors.Wrap(err, "could not create stream id")
	}

	s := &streamSession{
		id:      streamId,
		item:    it,
		seekPos: seekPos,
		dir:     c.streamDir(streamId),
		ready:   make(chan error),
		log:     c.log.New("trackId", fileId.String(), "conversionId", streamId.String()),
	}
	c.streams[streamId.String()] = s
	return s, nil
}

func (c *Converter) GetPlaylist(fileId domain.FileId, streamId domain.StreamId) (music.ConvertedFile, error) {
	s, err := c.getStreamForFile(streamId, fileId)
	if err != nil {
		return music.ConvertedFile{}, err
	}
	return openFile(c.playlistPathInDir(s.dir))
}

func (c *Converter) GetInit(fileId domain.FileId, streamId domain.StreamId) (music.ConvertedFile, error) {
	s, err := c.getStreamForFile(streamId, fileId)
	if err != nil {
		return music.ConvertedFile{}, err
	}
	return openFile(c.initPathInDir(s.dir))
}

func (c *Converter) GetFragment(fileId domain.FileId, streamId domain.StreamId, fragmentId domain.FragmentId) (music.ConvertedFile, error) {
	s, err := c.getStreamForFile(streamId, fileId)
	if err != nil {
		return music.ConvertedFile{}, err
	}
	return openFile(c.fragmentPathInDir(s.dir, fragmentId))
}

func (c *Converter) GetStats() (queries.StoreStats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var originalSize int64
	for _, it := range c.items {
		info, err := os.Stat(it.Path().String())
		if err != nil {
			return queries.StoreStats{}, errors.Wrap(err, "could not stat original")
		}
		originalSize += info.Size()
	}

	convertedSize, convertedCount, err := c.streamsStats()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not read stream stats")
	}

	return queries.StoreStats{
		AllItems:       int64(len(c.items)),
		ConvertedItems: convertedCount,
		OriginalSize:   originalSize,
		ConvertedSize:  convertedSize,
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

func (c *Converter) getStream(streamId domain.StreamId) (*streamSession, error) {
	c.mu.Lock()
	s, ok := c.streams[streamId.String()]
	c.mu.Unlock()
	if !ok {
		return nil, errors.New("stream does not exist")
	}
	return s, nil
}

func (c *Converter) getStreamForFile(streamId domain.StreamId, fileId domain.FileId) (*streamSession, error) {
	s, err := c.getStream(streamId)
	if err != nil {
		return nil, err
	}
	if s.item.FileId() != fileId {
		return nil, errors.New("stream does not belong to this track")
	}
	return s, nil
}

func (c *Converter) run(ctx context.Context, s *streamSession) {
	defer c.unregisterStream(s)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := os.MkdirAll(s.dir, 0755); err != nil {
		s.signalReady(ctx, errors.Wrap(err, "could not create the stream directory"))
		return
	}
	defer c.removeStreamDirectory(s)

	ffmpegResult := make(chan error, 1)
	select {
	case c.ffmpegJobs <- ffmpegJob{ctx: ctx, s: s, result: ffmpegResult}:
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

	<-ctx.Done()
}

func (c *Converter) waitForReadiness(ctx context.Context, s *streamSession, ffmpegResult <-chan error) error {
	playlist := c.playlistPathInDir(s.dir)
	initSegment := c.initPathInDir(s.dir)
	lastSegment := c.fragmentPathInDir(s.dir, readinessLastFragmentId)

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

func (c *Converter) unregisterStream(s *streamSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.streams, s.id.String())
}

func (c *Converter) removeStreamDirectory(s *streamSession) {
	if err := os.RemoveAll(s.dir); err != nil {
		s.log.Error("could not remove stream dir", "err", err, "dir", s.dir)
	}
}

func (s *streamSession) signalReady(ctx context.Context, err error) {
	select {
	case s.ready <- err:
	case <-ctx.Done():
	}
}

func (c *Converter) runFFmpeg(ctx context.Context, s *streamSession) error {
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
		"-hls_segment_filename", path.Join(s.dir, segmentTemplate),
		"-hls_base_url", "fragment/",
		c.playlistPathInDir(s.dir),
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

func mergeContexts(parent, other context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	stop := context.AfterFunc(other, cancel)
	return ctx, func() {
		stop()
		cancel()
	}
}

type streamSession struct {
	id      domain.StreamId
	item    music.TrackStoreItem
	seekPos *domain.SeekPosition
	dir     string
	log     logging.Logger

	ready chan error
}

type ffmpegJob struct {
	ctx    context.Context
	s      *streamSession
	result chan<- error
}
