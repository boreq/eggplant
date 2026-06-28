package http

import (
	"github.com/boreq/eggplant/adapters/openapi"
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/eggplant/domain/remote"
)

func toSearchResults(result library.SearchResults) openapi.SearchResults {
	return openapi.SearchResults{
		Albums: toAlbumSearchResults(result.Albums()),
		Tracks: toTrackWithAlbums(result.Tracks()),
	}
}

func toAlbumSearchResults(hits []library.FoundAlbum) []openapi.AlbumSearchResult {
	result := make([]openapi.AlbumSearchResult, 0, len(hits))
	for _, h := range hits {
		result = append(result, openapi.AlbumSearchResult{
			Album: toPartialAlbum(h.Album),
			Score: h.Dist,
		})
	}
	return result
}

func toTrackWithAlbums(hits []library.FoundTrack) []openapi.TrackWithAlbum {
	result := make([]openapi.TrackWithAlbum, 0, len(hits))
	for _, h := range hits {
		result = append(result, toTrackWithAlbum(h))
	}
	return result
}

func toTrackWithAlbum(h library.FoundTrack) openapi.TrackWithAlbum {
	var alb *openapi.PartialAlbum
	if a := h.Track.Album(); a != nil {
		v := toPartialAlbum(*a)
		alb = &v
	}
	return openapi.TrackWithAlbum{
		Track: toTrack(h.Track.Track()),
		Album: alb,
		Score: h.Dist,
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
		Id:               t.Id().String(),
		Number:           number,
		Title:            t.Title().String(),
		RemoteInstanceId: remoteInstanceId(t.RemoteInstanceId()),
	}
}

func remoteInstanceId(id *remote.RemoteInstanceID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func toAlbum(a music.Album) openapi.Album {
	return openapi.Album{
		Id:               a.Id().String(),
		Title:            a.Title().String(),
		Thumbnail:        toThumbnail(a.Thumbnail()),
		Parents:          toPartialAlbums(a.Parents()),
		Albums:           toPartialAlbums(a.Albums()),
		Tracks:           toTracks(a.Tracks().Items()),
		RemoteInstanceId: remoteInstanceId(a.RemoteInstanceId()),
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
		Id:               a.Id().String(),
		Title:            a.Title().String(),
		Thumbnail:        toThumbnail(a.Thumbnail()),
		RemoteInstanceId: remoteInstanceId(a.RemoteInstanceId()),
	}
}

func toTracks(tracks []music.Track) []openapi.Track {
	result := make([]openapi.Track, 0, len(tracks))
	for _, t := range tracks {
		result = append(result, toTrack(t))
	}
	return result
}
