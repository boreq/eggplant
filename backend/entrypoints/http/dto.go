package http

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
)

type searchResult struct {
	Albums []searchAlbum `json:"albums,omitempty"`
	Tracks []searchTrack `json:"tracks,omitempty"`
}

type searchAlbum struct {
	Path      []string   `json:"path,omitempty"`
	Title     string     `json:"title,omitempty"`
	Thumbnail *thumbnail `json:"thumbnail,omitempty"`
}

type thumbnail struct {
	Id string `json:"id,omitempty"`
}

type searchTrack struct {
	Track track             `json:"track,omitempty"`
	Album *searchTrackAlbum `json:"album,omitempty"`
}

type searchTrackAlbum struct {
	Path  []string `json:"path,omitempty"`
	Title string   `json:"title,omitempty"`
}

type track struct {
	Id       string  `json:"id,omitempty"`
	Title    string  `json:"title,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type album struct {
	Id        string     `json:"id,omitempty"`
	Title     string     `json:"title,omitempty"`
	Thumbnail *thumbnail `json:"thumbnail,omitempty"`
	Parents   []album    `json:"parents,omitempty"`
	Albums    []album    `json:"albums,omitempty"`
	Tracks    []track    `json:"tracks,omitempty"`
}

func toSearchResult(result library.SearchResult) searchResult {
	return searchResult{
		Albums: toSearchAlbums(result.Albums),
		Tracks: toSearchTracks(result.Tracks),
	}
}

func toSearchAlbums(albums []library.SearchAlbum) []searchAlbum {
	var result []searchAlbum
	for _, a := range albums {
		result = append(result, toSearchAlbum(a))
	}
	return result
}

func toSearchAlbum(a library.SearchAlbum) searchAlbum {
	return searchAlbum{
		Path:      toPath(a.Path),
		Title:     a.Title.String(),
		Thumbnail: toThumbnail(a.Thumbnail),
	}
}

func toPath(ids []domain.AlbumId) []string {
	var result []string
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}

func toThumbnail(thumb *domain.Thumbnail) *thumbnail {
	if thumb == nil {
		return nil
	}
	return &thumbnail{
		Id: thumb.Id().String(),
	}
}

func toSearchTracks(tracks []library.SearchTrack) []searchTrack {
	var result []searchTrack
	for _, t := range tracks {
		result = append(result, toSearchTrack(t))
	}
	return result
}

func toSearchTrack(t library.SearchTrack) searchTrack {
	return searchTrack{
		Track: toTrack(t.Track),
		Album: toSearchTrackAlbum(t.Album),
	}
}

func toSearchTrackAlbum(a *library.SearchTrackAlbum) *searchTrackAlbum {
	if a == nil {
		return nil
	}
	return &searchTrackAlbum{
		Path:  toPath(a.Path),
		Title: a.Title.String(),
	}
}

func toTrack(t domain.Track) track {
	return track{
		Id:       t.Id().String(),
		Title:    t.Title().String(),
		Duration: t.Duration().Seconds(),
	}
}

func toAlbum(a domain.Album) album {
	id := a.Id()
	title := a.Title()
	return album{
		Id:        id.String(),
		Title:     title.String(),
		Thumbnail: toThumbnail(a.Thumbnail()),
		Parents:   toParentAlbums(a.Parents()),
		Albums:    toAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks()),
	}
}

func toRootAlbum(a domain.RootAlbum) album {
	return album{
		Thumbnail: toThumbnail(a.Thumbnail()),
		Albums:    toAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks()),
	}
}

func toAlbums(albums []domain.ChildAlbum) []album {
	var result []album
	for _, a := range albums {
		id := a.Id()
		title := a.Title()
		result = append(result, album{
			Id:        id.String(),
			Title:     title.String(),
			Thumbnail: toThumbnail(a.Thumbnail()),
		})
	}
	return result
}

func toParentAlbums(parents []domain.ParentAlbum) []album {
	var result []album
	for _, p := range parents {
		id := p.Id()
		title := p.Title()
		result = append(result, album{
			Id:    id.String(),
			Title: title.String(),
		})
	}
	return result
}

func toTracks(tracks []domain.Track) []track {
	var result []track
	for _, t := range tracks {
		result = append(result, toTrack(t))
	}
	return result
}
