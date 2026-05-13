package library

import (
	"fmt"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type walkAlbumFn func(parent *music.BasicAlbum, id domain.AlbumId, v album) error

type walkTrackFn func(parent music.BasicAlbum, id domain.TrackId, v track) error

func (l *Library) walk(a walkAlbumFn, t walkTrackFn, publicOnly bool) error {
	access, err := l.getAccess(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get access")
	}

	if canAccess(access, publicOnly) {
		for id, track := range l.root.tracks {
			parent, err := newBasicAlbum(nil, *l.root)
			if err != nil {
				return errors.Wrap(err, "could not build basic album")
			}
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
	parentPath []domain.AlbumId,
	id domain.AlbumId,
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
		parent, err := newBasicAlbum(parentPath, *node)
		if err != nil {
			return errors.Wrap(err, "could not build basic album")
		}
		if err := a(&parent, id, *node); err != nil {
			return err
		}
	}

	if canAccess(access, publicOnly) {
		for id, track := range node.tracks {
			parent, err := newBasicAlbum(path, *node)
			if err != nil {
				return errors.Wrap(err, "could not build basic album")
			}
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

func newBasicAlbum(path []domain.AlbumId, album album) (music.BasicAlbum, error) {
	title, err := domain.NewAlbumTitle(album.title)
	if err != nil {
		return music.BasicAlbum{}, errors.Wrap(err, "invalid album title")
	}
	return music.BasicAlbum{
		Path:      path,
		Title:     title,
		Thumbnail: newThumbnail(album),
	}, nil
}

func newThumbnail(album album) *domain.Thumbnail {
	if album.thumbnailId == nil {
		return nil
	}
	t := domain.NewThumbnail(*album.thumbnailId)
	return &t
}
