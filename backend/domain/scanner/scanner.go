package scanner

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type FoundRootAlbum struct {
	thumbnailFile *domain.FilePath
	accessFile    *domain.FilePath
	albums        map[domain.AlbumTitle]FoundAlbum
	tracks        map[domain.TrackTitle]FoundTrack
}

func NewFoundRootAlbum(
	thumbnailFile *domain.FilePath,
	accessFile *domain.FilePath,
	albums map[domain.AlbumTitle]FoundAlbum,
	tracks map[domain.TrackTitle]FoundTrack,
) FoundRootAlbum {
	if albums == nil {
		albums = map[domain.AlbumTitle]FoundAlbum{}
	}
	if tracks == nil {
		tracks = map[domain.TrackTitle]FoundTrack{}
	}
	return FoundRootAlbum{
		thumbnailFile: thumbnailFile,
		accessFile:    accessFile,
		albums:        albums,
		tracks:        tracks,
	}
}

func (r FoundRootAlbum) ThumbnailFile() *domain.FilePath {
	return r.thumbnailFile
}

func (r FoundRootAlbum) AccessFile() *domain.FilePath {
	return r.accessFile
}

func (r FoundRootAlbum) Albums() map[domain.AlbumTitle]FoundAlbum {
	return r.albums
}

func (r FoundRootAlbum) Tracks() map[domain.TrackTitle]FoundTrack {
	return r.tracks
}

type FoundAlbum struct {
	thumbnailFile *domain.FilePath
	accessFile    *domain.FilePath
	albums        map[domain.AlbumTitle]FoundAlbum
	tracks        map[domain.TrackTitle]FoundTrack
}

func NewFoundAlbum(
	thumbnailFile *domain.FilePath,
	accessFile *domain.FilePath,
	albums map[domain.AlbumTitle]FoundAlbum,
	tracks map[domain.TrackTitle]FoundTrack,
) (FoundAlbum, error) {
	if len(albums) == 0 && len(tracks) == 0 {
		return FoundAlbum{}, errors.New("non-root album must have at least one track or sub-album")
	}
	if albums == nil {
		albums = map[domain.AlbumTitle]FoundAlbum{}
	}
	if tracks == nil {
		tracks = map[domain.TrackTitle]FoundTrack{}
	}
	return FoundAlbum{
		thumbnailFile: thumbnailFile,
		accessFile:    accessFile,
		albums:        albums,
		tracks:        tracks,
	}, nil
}

func (a FoundAlbum) ThumbnailFile() *domain.FilePath {
	return a.thumbnailFile
}

func (a FoundAlbum) AccessFile() *domain.FilePath {
	return a.accessFile
}

func (a FoundAlbum) Albums() map[domain.AlbumTitle]FoundAlbum {
	return a.albums
}

func (a FoundAlbum) Tracks() map[domain.TrackTitle]FoundTrack {
	return a.tracks
}

type FoundTrack struct {
	path domain.FilePath
}

func NewFoundTrack(path domain.FilePath) FoundTrack {
	return FoundTrack{
		path: path,
	}
}

func (t FoundTrack) Path() domain.FilePath {
	return t.path
}
