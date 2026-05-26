package scanner

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/errors"
)

type FoundRootAlbum struct {
	thumbnailFile *music.FilePath
	accessFile    *music.FilePath
	albums        map[music.AlbumTitle]FoundAlbum
	tracks        map[music.TrackTitle]FoundTrack
}

func NewFoundRootAlbum(
	thumbnailFile *music.FilePath,
	accessFile *music.FilePath,
	albums map[music.AlbumTitle]FoundAlbum,
	tracks map[music.TrackTitle]FoundTrack,
) FoundRootAlbum {
	if albums == nil {
		albums = map[music.AlbumTitle]FoundAlbum{}
	}
	if tracks == nil {
		tracks = map[music.TrackTitle]FoundTrack{}
	}
	return FoundRootAlbum{
		thumbnailFile: thumbnailFile,
		accessFile:    accessFile,
		albums:        albums,
		tracks:        tracks,
	}
}

func (r FoundRootAlbum) ThumbnailFile() *music.FilePath {
	return r.thumbnailFile
}

func (r FoundRootAlbum) AccessFile() *music.FilePath {
	return r.accessFile
}

func (r FoundRootAlbum) Albums() map[music.AlbumTitle]FoundAlbum {
	return r.albums
}

func (r FoundRootAlbum) Tracks() map[music.TrackTitle]FoundTrack {
	return r.tracks
}

type FoundAlbum struct {
	thumbnailFile *music.FilePath
	accessFile    *music.FilePath
	albums        map[music.AlbumTitle]FoundAlbum
	tracks        map[music.TrackTitle]FoundTrack
}

func NewFoundAlbum(
	thumbnailFile *music.FilePath,
	accessFile *music.FilePath,
	albums map[music.AlbumTitle]FoundAlbum,
	tracks map[music.TrackTitle]FoundTrack,
) (FoundAlbum, error) {
	if len(albums) == 0 && len(tracks) == 0 {
		return FoundAlbum{}, errors.New("non-root album must have at least one track or sub-album")
	}
	if albums == nil {
		albums = map[music.AlbumTitle]FoundAlbum{}
	}
	if tracks == nil {
		tracks = map[music.TrackTitle]FoundTrack{}
	}
	return FoundAlbum{
		thumbnailFile: thumbnailFile,
		accessFile:    accessFile,
		albums:        albums,
		tracks:        tracks,
	}, nil
}

func (a FoundAlbum) ThumbnailFile() *music.FilePath {
	return a.thumbnailFile
}

func (a FoundAlbum) AccessFile() *music.FilePath {
	return a.accessFile
}

func (a FoundAlbum) Albums() map[music.AlbumTitle]FoundAlbum {
	return a.albums
}

func (a FoundAlbum) Tracks() map[music.TrackTitle]FoundTrack {
	return a.tracks
}

type FoundTrack struct {
	path music.FilePath
}

func NewFoundTrack(path music.FilePath) FoundTrack {
	return FoundTrack{
		path: path,
	}
}

func (t FoundTrack) Path() music.FilePath {
	return t.path
}
