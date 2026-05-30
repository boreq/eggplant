package http

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/eggplant/entrypoints/http/openapi"
)

func toSearchResults(result library.SearchResults) openapi.SearchResults {
	return openapi.SearchResults{
		Albums: toPartialAlbums(result.Albums()),
		Tracks: toTrackWithAlbums(result.Tracks()),
	}
}

func toTrackWithAlbums(tracks []library.TrackWithAlbum) []openapi.TrackWithAlbum {
	result := make([]openapi.TrackWithAlbum, 0, len(tracks))
	for _, t := range tracks {
		result = append(result, toTrackWithAlbum(t))
	}
	return result
}

func toTrackWithAlbum(t library.TrackWithAlbum) openapi.TrackWithAlbum {
	var alb *openapi.PartialAlbum
	if a := t.Album(); a != nil {
		v := toPartialAlbum(*a)
		alb = &v
	}
	return openapi.TrackWithAlbum{
		Track: toTrack(t.Track()),
		Album: alb,
	}
}

func toThumbnail(thumb *music.Thumbnail) *openapi.Thumbnail {
	if thumb == nil {
		return nil
	}
	return &openapi.Thumbnail{
		Id: thumb.Id().String(),
	}
}

func toTrack(t music.Track) openapi.Track {
	var number *int
	if n := t.Number(); n != nil {
		v := n.Int()
		number = &v
	}
	return openapi.Track{
		Id:     t.Id().String(),
		Number: number,
		Title:  t.Title().String(),
	}
}

func toAlbum(a music.Album) openapi.Album {
	return openapi.Album{
		Id:        a.Id().String(),
		Title:     a.Title().String(),
		Thumbnail: toThumbnail(a.Thumbnail()),
		Parents:   toPartialAlbums(a.Parents()),
		Albums:    toPartialAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks().Items()),
	}
}

func toRootAlbum(a music.RootAlbum) openapi.RootAlbum {
	return openapi.RootAlbum{
		Thumbnail: toThumbnail(a.Thumbnail()),
		Albums:    toPartialAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks().Items()),
	}
}

func toPartialAlbums(albums []music.PartialAlbum) []openapi.PartialAlbum {
	result := make([]openapi.PartialAlbum, 0, len(albums))
	for _, a := range albums {
		result = append(result, toPartialAlbum(a))
	}
	return result
}

func toPartialAlbum(a music.PartialAlbum) openapi.PartialAlbum {
	return openapi.PartialAlbum{
		Id:        a.Id().String(),
		Title:     a.Title().String(),
		Thumbnail: toThumbnail(a.Thumbnail()),
	}
}

func toTracks(tracks []music.Track) []openapi.Track {
	result := make([]openapi.Track, 0, len(tracks))
	for _, t := range tracks {
		result = append(result, toTrack(t))
	}
	return result
}
