package http

import (
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
)

type searchResult struct {
	Albums []basicAlbum        `json:"albums,omitempty"`
	Tracks []searchResultTrack `json:"tracks,omitempty"`
}

type basicAlbum struct {
	Path      []string   `json:"path,omitempty"`
	Title     string     `json:"title,omitempty"`
	Thumbnail *thumbnail `json:"thumbnail,omitempty"`
}

type thumbnail struct {
	FileId string `json:"fileId,omitempty"`
}

type searchResultTrack struct {
	Track track      `json:"track,omitempty"`
	Album basicAlbum `json:"album,omitempty"`
}

type track struct {
	Id       string  `json:"id,omitempty"`
	FileId   string  `json:"fileId,omitempty"`
	Title    string  `json:"title,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

type album struct {
	Id        string     `json:"id,omitempty"`
	Title     string     `json:"title,omitempty"`
	Thumbnail *thumbnail `json:"thumbnail,omitempty"`
	Access    access     `json:"access,omitempty"`
	Parents   []album    `json:"parents,omitempty"`
	Albums    []album    `json:"albums,omitempty"`
	Tracks    []track    `json:"tracks,omitempty"`
}

type access struct {
	Public bool `json:"public"`
}

func toSearchResult(result music.SearchResult) searchResult {
	return searchResult{
		Albums: toBasicAlbums(result.Albums),
		Tracks: toSearchResultTracks(result.Tracks),
	}
}

func toBasicAlbums(albums []music.BasicAlbum) []basicAlbum {
	var result []basicAlbum
	for _, a := range albums {
		result = append(result, toBasicAlbum(a))
	}
	return result
}

func toBasicAlbum(a music.BasicAlbum) basicAlbum {
	return basicAlbum{
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
	fileId := thumb.FileId()
	return &thumbnail{
		FileId: fileId.String(),
	}
}

func toSearchResultTracks(tracks []music.SearchResultTrack) []searchResultTrack {
	var result []searchResultTrack
	for _, t := range tracks {
		result = append(result, toSearchResultTrack(t))
	}
	return result
}

func toSearchResultTrack(t music.SearchResultTrack) searchResultTrack {
	return searchResultTrack{
		Track: toTrack(t.Track),
		Album: toBasicAlbum(t.Album),
	}
}

func toTrack(t domain.Track) track {
	var duration float64
	if d := t.Duration(); d != nil {
		duration = d.Seconds()
	}
	id := t.Id()
	fileId := t.FileId()
	title := t.Title()
	return track{
		Id:       id.String(),
		FileId:   fileId.String(),
		Title:    title.String(),
		Duration: duration,
	}
}

func toAlbum(a domain.Album) album {
	id := a.Id()
	title := a.Title()
	return album{
		Id:        id.String(),
		Title:     title.String(),
		Thumbnail: toThumbnail(a.Thumbnail()),
		Access:    toAccess(a.Access()),
		Parents:   toAlbumParents(a.Parents()),
		Albums:    toAlbums(a.Albums()),
		Tracks:    toTracks(a.Tracks()),
	}
}

func toAlbums(albums []domain.Album) []album {
	var result []album
	for _, a := range albums {
		result = append(result, toAlbum(a))
	}
	return result
}

func toAlbumParents(parents []domain.AlbumParent) []album {
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

func toAccess(a domain.Access) access {
	return access{
		Public: a.Public(),
	}
}
