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
	initFilename     = "init.mp4"
	segmentTemplate  = "seg_%04d.m4s"
	trackDirectory   = "tracks"
	playlistEndTag   = "#EXT-X-ENDLIST"

	hlsSegmentSeconds = 4
	readinessTimeout  = 60 * time.Second
	readinessPoll     = 100 * time.Millisecond

	cacheItemsFor     = 30 * time.Minute
	cleanupEvery      = cacheItemsFor / 2
	cleanupErrorDelay = 1 * time.Minute
)

var firstFragmentId = domain.MustNewTrackFragmentId(0)

type Converter struct {
	dataDir string
	log     logging.Logger

	mu               sync.Mutex
	items            map[domain.FileId]item
	itemsAccessTimes map[domain.FileId]time.Time
	ongoing          map[domain.FileId]*conversion

	workCh chan *conversion
}

func NewConverter(ctx context.Context, dataDir string) (*Converter, error) {
	c := &Converter{
		dataDir:          dataDir,
		log:              logging.New("tracks.Converter"),
		items:            make(map[domain.FileId]item),
		itemsAccessTimes: make(map[domain.FileId]time.Time),
		ongoing:          make(map[domain.FileId]*conversion),
		workCh:           make(chan *conversion),
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go c.worker(ctx)
	}
	go c.cleanupLoop(ctx)

	return c, nil
}

func (c *Converter) SetItems(items []music.TrackStoreItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[domain.FileId]item)
	for _, it := range items {
		c.items[it.FileId()] = item{id: it.FileId(), path: it.Path()}
	}

	for id := range c.itemsAccessTimes {
		if _, ok := c.items[id]; !ok {
			delete(c.itemsAccessTimes, id)
		}
	}
}

func (c *Converter) GetPlaylist(ctx context.Context, fileId domain.FileId) (domain.ConvertedFile, error) {
	if err := c.ensureConverted(ctx, fileId); err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "error requesting conversion")
	}
	return openFile(c.playlistPath(fileId))
}

func (c *Converter) GetInit(ctx context.Context, fileId domain.FileId) (domain.ConvertedFile, error) {
	if err := c.ensureConverted(ctx, fileId); err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "error requesting conversion")
	}
	return openFile(c.initPath(fileId))
}

func (c *Converter) GetFragment(ctx context.Context, fileId domain.FileId, fragmentId domain.TrackFragmentId) (domain.ConvertedFile, error) {
	if err := c.ensureConverted(ctx, fileId); err != nil {
		return domain.ConvertedFile{}, errors.Wrap(err, "error requesting conversion")
	}
	return openFile(c.fragmentPath(fileId, fragmentId))
}

func (c *Converter) GetStats() (queries.StoreStats, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var originalSize int64
	for _, it := range c.items {
		info, err := os.Stat(it.path.String())
		if err != nil {
			return queries.StoreStats{}, errors.Wrap(err, "could not stat original")
		}
		originalSize += info.Size()
	}

	convertedSize, convertedCount, err := c.convertedStats()
	if err != nil {
		return queries.StoreStats{}, errors.Wrap(err, "could not read converted stats")
	}

	return queries.StoreStats{
		AllItems:       int64(len(c.items)),
		ConvertedItems: convertedCount,
		OriginalSize:   originalSize,
		ConvertedSize:  convertedSize,
	}, nil
}

func (c *Converter) convertedStats() (size int64, count int64, err error) {
	entries, err := os.ReadDir(c.outputRoot())
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
		subEntries, err := os.ReadDir(path.Join(c.outputRoot(), entry.Name()))
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

func (c *Converter) ensureConverted(ctx context.Context, fileId domain.FileId) error {
	c.mu.Lock()
	if _, ok := c.items[fileId]; !ok {
		c.mu.Unlock()
		return errors.New("item does not exist")
	}
	c.itemsAccessTimes[fileId] = time.Now()
	c.mu.Unlock()

	conv := &conversion{
		fileId: fileId,
		done:   make(chan struct{}),
	}

	select {
	case c.workCh <- conv:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-conv.done:
		return conv.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Converter) worker(ctx context.Context) {
	for {
		select {
		case conv := <-c.workCh:
			c.runConversion(ctx, conv)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Converter) runConversion(ctx context.Context, conv *conversion) {
	// if another worker is already running this conversion wait for it and inherit its outcome
	c.mu.Lock()
	if alreadyOngoing, ok := c.ongoing[conv.fileId]; ok {
		c.mu.Unlock()
		<-alreadyOngoing.done
		conv.finishedWithErr(alreadyOngoing.err)
		return
	}

	c.ongoing[conv.fileId] = conv
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.ongoing, conv.fileId)
	}()

	ok, err := c.verifyCompleted(conv.fileId)
	if err != nil {
		conv.finishedWithErr(errors.Wrap(err, "error checking if file is complete"))
		return
	}
	if ok {
		conv.finishedWithErr(nil)
		return
	}

	c.mu.Lock()
	it, ok := c.items[conv.fileId]
	c.mu.Unlock()
	if !ok {
		conv.finishedWithErr(errors.New("item does not exist"))
		return
	}

	start := time.Now()
	err = c.runFFmpeg(ctx, conv, it)
	if err != nil {
		if removeErr := os.RemoveAll(c.trackDir(conv.fileId)); removeErr != nil {
			c.log.Error("could not clean up after failed conversion", "err", removeErr)
		}
	}
	conv.finishedWithErr(err)
	c.log.Debug("conversion ended", "err", err, "duration", time.Since(start))
}

func (c *Converter) runFFmpeg(ctx context.Context, conv *conversion, it item) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd, stderr, err := c.startFFmpeg(ctx, it)
	if err != nil {
		return errors.Wrap(err, "error starting ffmpeg")
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ffmpegDone, err := c.waitForReadiness(ctx, it.id, cmd, stderr, done)
	if err != nil {
		return errors.Wrap(err, "error waiting for ffmpeg readiness")
	}

	// Readiness reached; let the caller proceed even though ffmpeg may still
	// be churning out later segments.
	conv.finishedWithErr(nil)

	if ffmpegDone {
		return nil
	}
	return c.waitForExit(ctx, stderr, done)
}

func (c *Converter) startFFmpeg(ctx context.Context, it item) (*exec.Cmd, *bytes.Buffer, error) {
	dir := c.trackDir(it.id)
	if err := os.RemoveAll(dir); err != nil {
		return nil, nil, errors.Wrap(err, "could not clear the per-track output directory")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, errors.Wrap(err, "could not create the per-track output directory")
	}

	args := []string{
		"-y",
		"-i", it.path.String(),
		"-vn",
		"-c:a", "libopus",
		"-b:a", "96K",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", hlsSegmentSeconds),
		"-hls_playlist_type", "event",
		"-hls_segment_type", "fmp4",
		"-hls_fmp4_init_filename", initFilename,
		"-hls_segment_filename", path.Join(dir, segmentTemplate),
		"-hls_base_url", "fragment/",
		c.playlistPath(it.id),
	}
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	c.log.Debug("converting", "command", cmd.String())

	if err := cmd.Start(); err != nil {
		return nil, nil, errors.Wrap(err, "ffmpeg start failed")
	}
	return cmd, stderr, nil
}

// waitForReadiness blocks until the playlist, init, and first segment exist
// on disk, ffmpeg exits, the timeout fires, or ctx is canceled. The first
// return value reports whether ffmpeg has already exited (and so `done` has
// been drained).
func (c *Converter) waitForReadiness(ctx context.Context, fileId domain.FileId, cmd *exec.Cmd, stderr *bytes.Buffer, done <-chan error) (bool, error) {
	playlist := c.playlistPath(fileId)
	firstSegment := c.fragmentPath(fileId, firstFragmentId)
	initSegment := c.initPath(fileId)

	tick := time.NewTicker(readinessPoll)
	defer tick.Stop()
	timeout := time.After(readinessTimeout)

	for {
		select {
		case err := <-done:
			if err != nil {
				return true, errors.Wrapf(err, "ffmpeg failed (stderr: %s)", strings.TrimSpace(stderr.String()))
			}
			return true, nil
		case <-timeout:
			cmd.Process.Kill()
			<-done
			return true, errors.New("timed out waiting for ffmpeg to produce initial output")
		case <-tick.C:
			if fileExists(playlist) && fileExists(firstSegment) && fileExists(initSegment) {
				return false, nil
			}
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
}

// waitForExit blocks until ffmpeg fully exits. The caller has already been
// signaled at readiness; the error here can't change that, but it's still
// surfaced so the conversion log line reflects what actually happened.
func (c *Converter) waitForExit(ctx context.Context, stderr *bytes.Buffer, done <-chan error) error {
	select {
	case err := <-done:
		if err != nil {
			return errors.Wrapf(err, "ffmpeg failed after readiness (stderr: %s)", strings.TrimSpace(stderr.String()))
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Converter) cleanupLoop(ctx context.Context) {
	for {
		if err := c.cleanup(); err != nil {
			c.log.Error("cleanup failed", "err", err)
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

func (c *Converter) cleanup() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.log.Debug("performing cleanup", "dir", c.outputRoot())

	canExist := make(map[string]struct{})
	for id, t := range c.itemsAccessTimes {
		if time.Since(t) < cacheItemsFor {
			canExist[c.trackDir(id)] = struct{}{}
		}
	}
	for id := range c.ongoing {
		canExist[c.trackDir(id)] = struct{}{}
	}

	entries, err := os.ReadDir(c.outputRoot())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "could not read the output directory")
	}
	for _, entry := range entries {
		p := path.Join(c.outputRoot(), entry.Name())
		if _, ok := canExist[p]; !ok {
			c.log.Debug("removing", "path", p)
			if err := os.RemoveAll(p); err != nil {
				return errors.Wrap(err, "remove all error")
			}
		}
	}
	return nil
}

func (c *Converter) outputRoot() string {
	return path.Join(c.dataDir, trackDirectory)
}

func (c *Converter) trackDir(id domain.FileId) string {
	return path.Join(c.outputRoot(), id.String())
}

func (c *Converter) playlistPath(id domain.FileId) string {
	return path.Join(c.trackDir(id), playlistFilename)
}

func (c *Converter) initPath(id domain.FileId) string {
	return path.Join(c.trackDir(id), initFilename)
}

func (c *Converter) fragmentPath(id domain.FileId, n domain.TrackFragmentId) string {
	return path.Join(c.trackDir(id), fmt.Sprintf(segmentTemplate, n.Int()))
}

func (c *Converter) verifyCompleted(fileId domain.FileId) (bool, error) {
	f, err := os.Open(c.playlistPath(fileId))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, errors.Wrap(err, "could not open the playlist")
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false, errors.Wrap(err, "could not stat the playlist")
	}

	const tailSize = 64
	size := info.Size()
	if size > tailSize {
		size = tailSize
	}
	buf := make([]byte, size)
	if _, err := f.ReadAt(buf, info.Size()-size); err != nil {
		return false, errors.Wrap(err, "could not read the playlist tail")
	}

	return bytes.HasSuffix(bytes.TrimRight(buf, " \t\r\n"), []byte(playlistEndTag)), nil
}

func openFile(p string) (domain.ConvertedFile, error) {
	f, err := os.Open(p)
	if err != nil {
		return domain.ConvertedFile{}, errors.Wrapf(err, "could not open '%s'", p)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return domain.ConvertedFile{}, errors.Wrap(err, "stat failed")
	}
	return domain.ConvertedFile{
		Name:    f.Name(),
		Modtime: info.ModTime(),
		Content: f,
	}, nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

type item struct {
	id   domain.FileId
	path domain.FilePath
}

type conversion struct {
	fileId domain.FileId
	done   chan struct{}
	err    error
	once   sync.Once
}

// finishedWithErr signals waiters with the given outcome. The first call wins;
// subsequent calls are no-ops, so it's safe to call from both happy paths and
// deferred error handlers without worrying about double-close.
func (c *conversion) finishedWithErr(err error) {
	c.once.Do(func() {
		c.err = err
		close(c.done)
	})
}
