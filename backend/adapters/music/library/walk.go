package library

import (
	"fmt"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/errors"
)

type walkAlbumFn func(parent *music.BasicAlbum, id music.AlbumId, v album) error

type walkTrackFn func(parent music.BasicAlbum, id music.TrackId, v track) error

func (l *Library) walk(a walkAlbumFn, t walkTrackFn, publicOnly bool) error {
	access, err := l.getAccess(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get access")
	}

	if canAccess(access, publicOnly) {
		for id, track := range l.root.tracks {
			parent := newBasicAlbum(nil, *l.root)
			if err := t(parent, id, track); err != nil {
				return err
			}
		}
	}

	for id, album := range l.root.albums {
		if err := l.subWalk(nil, id, album, a, t, publicOnly); err != nil {
			return err
		}
	}

	return nil
}

func (l *Library) subWalk(
	parentPath []music.AlbumId,
	id music.AlbumId,
	node *album,
	a walkAlbumFn,
	t walkTrackFn,
	publicOnly bool,
) error {
	path := append(
		parentPath,
		id,
	)

	access, err := l.getAccess(path)
	if err != nil {
		return errors.Wrap(err, "failed to get access")
	}

	fmt.Println(access)

	if canAccess(access, publicOnly) {
		parent := newBasicAlbum(parentPath, *node)
		if err := a(&parent, id, *node); err != nil {
			return err
		}
	}

	if canAccess(access, publicOnly) {
		for id, track := range node.tracks {
			parent := newBasicAlbum(path, *node)
			if err := t(parent, id, track); err != nil {
				return err
			}
		}
	}

	for id, childAlbum := range node.albums {
		if err := l.subWalk(path, id, childAlbum, a, t, publicOnly); err != nil {
			return err
		}
	}

	return nil
}

func newBasicAlbum(path []music.AlbumId, album album) music.BasicAlbum {
	return music.BasicAlbum{
		Path:      path,
		Title:     album.title,
		Thumbnail: newThumbnail(album),
	}
}

func newThumbnail(album album) *music.Thumbnail {
	if album.thumbnailId != "" {
		return &music.Thumbnail{
			FileId: album.thumbnailId,
		}
	}
	return nil
}
