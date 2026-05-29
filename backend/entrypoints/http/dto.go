package http

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
)

type searchResults struct {
	Albums []partialAlbum   `json:"albums"`
	Tracks []trackWithAlbum `json:"tracks"`
}

type trackWithAlbum struct {
	Track track         `json:"track"`
	Album *partialAlbum `json:"album,omitempty"`
}

type rootAlbum struct {
	Thumbnail *thumbnail     `json:"thumbnail,omitempty"`
	Albums    []partialAlbum `json:"albums"`
	Tracks    []track        `json:"tracks"`
}

type album struct {
	Id        string         `json:"id"`
	Title     string         `json:"title"`
	Thumbnail *thumbnail     `json:"thumbnail,omitempty"`
	Parents   []partialAlbum `json:"parents"`
	Albums    []partialAlbum `json:"albums"`
	Tracks    []track        `json:"tracks"`
}

type partialAlbum struct {
	Id        string     `json:"id"`
	Title     string     `json:"title"`
	Thumbnail *thumbnail `json:"thumbnail,omitempty"`
}

type thumbnail struct {
	Id string `json:"id"`
}

type track struct {
	Id     string `json:"id"`
	Number *int   `json:"number,omitempty"`
	Title  string `json:"title"`
}

type trackDuration struct {
	Duration float64 `json:"duration"`
}

func toSearchResults(result library.SearchResults) searchResults {
	return searchResults{
		Albums: toPartialAlbums(result.Albums()),
		Tracks: toTrackWithAlbums(result.Tracks()),
	}
}

func toTrackWithAlbums(tracks []library.TrackWithAlbum) []trackWithAlbum {
	result := make([]trackWithAlbum, 0, len(tracks))
	for _, t := range tracks {
		result = append(result, toTrackWithAlbum(t))
	}
	return result
}

func toTrackWithAlbum(t library.TrackWithAlbum) trackWithAlbum {
	var alb *partialAlbum
	if a := t.Album(); a != nil {
		v := toPartialAlbum(*a)
		alb = &v
	}
	return trackWithAlbum{
		Track: toTrack(t.Track()),
		Album: alb,
	}
}

func toThumbnail(thumb *music.Thumbnail) *thumbnail {
	if thumb == nil {
		return nil
	}
	return &thumbnail{
		Id: thumb.Id().String(),
	}
}

func toTrack(t music.Track) track {
	var number *int
	if n := t.Number(); n != nil {
		v := n.Int()
		number = &v
	}
	return track{
		Id:     t.Id().String(),
		Number: number,
		Title:  t.Title().String(),
	}
}

func toAlbum(a music.Album) album {
	return album{
		Id:        a.Id().String(),
		Title:     a.Title().String(),
		Thumbnail: toThumbnail(a.Thumbnail()),
		Parents:   toPartialAlbums(a.Parents()),
		Albums:    toPartialAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks().Items()),
	}
}

func toRootAlbum(a music.RootAlbum) rootAlbum {
	return rootAlbum{
		Thumbnail: toThumbnail(a.Thumbnail()),
		Albums:    toPartialAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks().Items()),
	}
}

func toPartialAlbums(albums []music.PartialAlbum) []partialAlbum {
	result := make([]partialAlbum, 0, len(albums))
	for _, a := range albums {
		result = append(result, toPartialAlbum(a))
	}
	return result
}

func toPartialAlbum(a music.PartialAlbum) partialAlbum {
	return partialAlbum{
		Id:        a.Id().String(),
		Title:     a.Title().String(),
		Thumbnail: toThumbnail(a.Thumbnail()),
	}
}

func toTracks(tracks []music.Track) []track {
	result := make([]track, 0, len(tracks))
	for _, t := range tracks {
		result = append(result, toTrack(t))
	}
	return result
}
