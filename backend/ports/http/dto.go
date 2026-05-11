package http

import "github.com/boreq/eggplant/application/music"

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

func toSearchResult(result music.SearchResult) searchResult {
	return searchResult{
		Albums: toBasicAlbums(result.Albums),
		Tracks: toSearchResultTracks(
			result.Tracks,
		),
	}
}

func toBasicAlbums(albums []music.BasicAlbum) []basicAlbum {
	var result []basicAlbum
	for _, album := range albums {
		result = append(
			result,
			toBasicAlbum(album),
		)
	}
	return result
}

func toBasicAlbum(album music.BasicAlbum) basicAlbum {
	return basicAlbum{
		Path:      toPath(album.Path),
		Title:     album.Title,
		Thumbnail: toThumbnail(album.Thumbnail),
	}
}

func toPath(ids []music.AlbumId) []string {
	var result []string
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}

func toThumbnail(thumb *music.Thumbnail) *thumbnail {
	if thumb == nil {
		return nil
	}

	return &thumbnail{
		FileId: thumb.FileId.String(),
	}
}

func toSearchResultTracks(tracks []music.SearchResultTrack) []searchResultTrack {

	var result []searchResultTrack
	for _, track := range tracks {
		result = append(
			result,
			toSearchResultTrack(track),
		)
	}
	return result
}

func toSearchResultTrack(track music.SearchResultTrack) searchResultTrack {
	return searchResultTrack{
		Track: toTrack(track.Track),
		Album: toBasicAlbum(track.Album),
	}
}

func toTrack(t music.Track) track {
	return track{
		Id:       t.Id.String(),
		FileId:   t.FileId.String(),
		Title:    t.Title,
		Duration: t.Duration,
	}
}
