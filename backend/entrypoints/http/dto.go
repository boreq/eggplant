package http

import (
	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/library"
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
	Parents   []album    `json:"parents,omitempty"`
	Albums    []album    `json:"albums,omitempty"`
	Tracks    []track    `json:"tracks,omitempty"`
}

func toSearchResult(result library.SearchResult) searchResult {
	return searchResult{
		Albums: toBasicAlbums(result.Albums),
		Tracks: toSearchResultTracks(result.Tracks),
	}
}

func toBasicAlbums(albums []library.BasicAlbum) []basicAlbum {
	var result []basicAlbum
	for _, a := range albums {
		result = append(result, toBasicAlbum(a))
	}
	return result
}

func toBasicAlbum(a library.BasicAlbum) basicAlbum {
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
	id := thumb.Id()
	return &thumbnail{
		FileId: id.String(),
	}
}

func toSearchResultTracks(tracks []library.SearchResultTrack) []searchResultTrack {
	var result []searchResultTrack
	for _, t := range tracks {
		result = append(result, toSearchResultTrack(t))
	}
	return result
}

func toSearchResultTrack(t library.SearchResultTrack) searchResultTrack {
	return searchResultTrack{
		Track: toTrack(t.Track),
		Album: toBasicAlbum(t.Album),
	}
}

func toTrack(t domain.Track) track {
	id := t.Id()
	title := t.Title()
	return track{
		Id:       id.String(),
		FileId:   id.String(),
		Title:    title.String(),
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
