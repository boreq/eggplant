package store

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/boreq/eggplant/logging"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

const scanEvery = 10 * time.Second

type Thumbnail struct {
	Id   string
	Path string
}

func NewThumbnailStore(cacheDir string) (*ThumbnailStore, error) {
	s := &ThumbnailStore{
		cacheDir: cacheDir,
		ch:       make(chan []Thumbnail),
		log:      logging.New("thumbnailStore"),
	}
	go s.receiveThumbnails()
	go s.processThumbnails()
	return s, nil
}

type ThumbnailStore struct {
	cacheDir      string
	ch            chan []Thumbnail
	thumbnails    []Thumbnail
	thumbnailsSet bool
	mutex         sync.Mutex
	log           logging.Logger
}

func (s *ThumbnailStore) SetThumbnails(thumbnails []Thumbnail) {
	s.ch <- thumbnails
}

func (s *ThumbnailStore) ServeFile(w http.ResponseWriter, r *http.Request, id string) {
	http.ServeFile(w, r, s.filePath(id))
}

func (s *ThumbnailStore) receiveThumbnails() {
	for thumbnails := range s.ch {
		if err := s.handleThumbnails(thumbnails); err != nil {
			s.log.Error("could not handle updates", "err", err)
		}
	}
}

func (s *ThumbnailStore) handleThumbnails(thumbnails []Thumbnail) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.thumbnails = thumbnails
	s.thumbnailsSet = true
	return nil
}

func (s *ThumbnailStore) processThumbnails() {
	for {
		thumbnail, ok := s.getNextThumbnail()
		if !ok {
			s.log.Debug("no thumbnails to convert")
			<-time.After(scanEvery)
			continue
		} else {
			s.log.Debug("converting a thumbnail", "thumbnail", thumbnail)
			if err := s.convert(thumbnail); err != nil {
				s.log.Error("conversion failed", "err", err)
				<-time.After(time.Second)
			}
		}
	}
}

func (s *ThumbnailStore) getNextThumbnail() (Thumbnail, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	rand.Shuffle(len(s.thumbnails), func(i, j int) {
		s.thumbnails[i], s.thumbnails[j] = s.thumbnails[j], s.thumbnails[i]
	})

	for _, thumbnail := range s.thumbnails {
		p := s.filePath(thumbnail.Id)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return thumbnail, true
		}
	}
	return Thumbnail{}, false
}

func (s *ThumbnailStore) convert(thumbnail Thumbnail) error {
	outputPath := s.filePath(thumbnail.Id)
	tmpOutputPath := s.tmpFilePath(thumbnail.Id)

	if err := makeDirectory(outputPath); err != nil {
		return errors.Wrap(err, "could not create the output directory")
	}

	f, err := os.Open(thumbnail.Path)
	if err != nil {
		return errors.Wrap(err, "could not open the input file")
	}
	defer f.Close()

	output, err := os.Create(tmpOutputPath)
	if err != nil {
		return errors.Wrap(err, "could not create an output file")
	}
	defer output.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return errors.Wrap(err, "decoding failed")
	}

	resized := resize.Resize(thumbnailSize, thumbnailSize, img, resize.Lanczos3)

	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(output, resized); err != nil {
		return errors.Wrap(err, "encoding failed")
	}

	if err := os.Rename(tmpOutputPath, outputPath); err != nil {
		return errors.Wrap(err, "move failed")
	}

	return nil
}

const thumbnailSize = 200
const thumbnailExtension = "png"
const thumbnailDirectory = "thumbnails"

func (s *ThumbnailStore) filePath(id string) string {
	dir := path.Join(s.cacheDir, thumbnailDirectory)
	file := fmt.Sprintf("%s.%s", id, thumbnailExtension)
	return path.Join(dir, file)
}

func (s *ThumbnailStore) tmpFilePath(id string) string {
	dir := path.Join(s.cacheDir, thumbnailDirectory)
	file := fmt.Sprintf("_%s.%s", id, thumbnailExtension)
	return path.Join(dir, file)
}

func makeDirectory(file string) error {
	dir, _ := path.Split(file)
	return os.MkdirAll(dir, os.ModePerm)
}
