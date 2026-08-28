package library

import (
	"slices"
	"strings"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
)

const maxSearchItems = 20

func (l *Library) Search(accessCtx AccessContext, query string) (SearchResults, error) {
	builder := newSearchBuilder()

	if accessCtx.CanSee(l.getRootVisibility()) {
		for _, track := range l.root.tracks {
			if !containsStringCaseInsensitive(track.Title().String(), query) {
				continue
			}
			builder.addTrack(NewRootTrackWithAlbum(track), 0)
		}
	}

	searchAlbums(builder, l.root.albums, nil, l.getRootVisibilityToPropagateToChildren(), query, accessCtx)
	return builder.build(), nil
}

func searchAlbums(b *searchBuilder, albums []Album, parentInheritedDist *int, parentVis Visibility, query string, ctx AccessContext) {
	for _, a := range albums {
		effective := parentVis
		if a.visibility != nil {
			effective = *a.visibility
		}

		var ownDist *int
		if containsStringCaseInsensitive(a.title.String(), query) {
			ownDist = new(0)
		} else {
			ownDist = parentInheritedDist
		}

		if ctx.CanSee(effective) {
			album := music.NewPartialAlbum(a.id, a.title, a.thumbnail)

			if ownDist != nil {
				b.addAlbum(album, *ownDist)
			}

			childDist := incDist(ownDist)
			for _, track := range a.tracks {
				if containsStringCaseInsensitive(track.Title().String(), query) {
					b.addTrack(NewTrackWithAlbum(track, album), 0)
				} else if childDist != nil {
					b.addTrack(NewTrackWithAlbum(track, album), *childDist)
				}
			}
		}

		searchAlbums(b, a.albums, incDist(ownDist), effective, query, ctx)
	}
}

type FoundAlbum struct {
	Album music.PartialAlbum
	Dist  int
}

type FoundTrack struct {
	Track TrackWithAlbum
	Dist  int
}

type SearchResults struct {
	albums []FoundAlbum
	tracks []FoundTrack
}

func NewSearchResults(albums []FoundAlbum, tracks []FoundTrack) SearchResults {
	return SearchResults{albums: albums, tracks: tracks}
}

func (s SearchResults) Albums() []FoundAlbum {
	return s.albums
}

func (s SearchResults) Tracks() []FoundTrack {
	return s.tracks
}

func MergeResults(results ...SearchResults) SearchResults {
	b := newSearchBuilder()
	for _, r := range results {
		for _, h := range r.albums {
			b.addAlbum(h.Album, h.Dist)
		}
		for _, h := range r.tracks {
			b.addTrack(h.Track, h.Dist)
		}
	}
	return b.build()
}

type TrackWithAlbum struct {
	track music.Track
	album *music.PartialAlbum
}

func NewRootTrackWithAlbum(track music.Track) TrackWithAlbum {
	return TrackWithAlbum{track: track, album: nil}
}

func NewTrackWithAlbum(track music.Track, album music.PartialAlbum) TrackWithAlbum {
	return TrackWithAlbum{track: track, album: &album}
}

func (t TrackWithAlbum) Track() music.Track {
	return t.track
}

func (t TrackWithAlbum) Album() *music.PartialAlbum {
	return t.album
}

func incDist(d *int) *int {
	if d == nil {
		return nil
	}
	return new(*d + 1)
}

func containsStringCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type albumHit struct {
	album music.PartialAlbum
	dist  int
}

type trackHit struct {
	track TrackWithAlbum
	dist  int
}

type searchBuilder struct {
	albums map[string]albumHit
	tracks map[string]trackHit
}

func newSearchBuilder() *searchBuilder {
	return &searchBuilder{
		albums: map[string]albumHit{},
		tracks: map[string]trackHit{},
	}
}

func (b *searchBuilder) addAlbum(a music.PartialAlbum, dist int) {
	key := instanceKey(a.RemoteInstanceId()) + a.Id().String()
	if existing, ok := b.albums[key]; ok && existing.dist <= dist {
		return
	}
	b.albums[key] = albumHit{album: a, dist: dist}
}

func (b *searchBuilder) addTrack(t TrackWithAlbum, dist int) {
	key := instanceKey(t.Track().RemoteInstanceId()) + t.Track().Id().String()
	if existing, ok := b.tracks[key]; ok && existing.dist <= dist {
		return
	}
	b.tracks[key] = trackHit{track: t, dist: dist}
}

func instanceKey(id *remote.RemoteInstanceID) string {
	if id == nil {
		return ""
	}
	return id.String() + "/"
}

func (b *searchBuilder) build() SearchResults {
	albumHits := make([]albumHit, 0, len(b.albums))
	for _, h := range b.albums {
		albumHits = append(albumHits, h)
	}
	slices.SortStableFunc(albumHits, func(a, b albumHit) int {
		if a.dist != b.dist {
			if a.dist < b.dist {
				return -1
			}
			return 1
		}
		return strings.Compare(a.album.Title().String(), b.album.Title().String())
	})

	trackHits := make([]trackHit, 0, len(b.tracks))
	for _, h := range b.tracks {
		trackHits = append(trackHits, h)
	}
	slices.SortStableFunc(trackHits, func(a, b trackHit) int {
		if a.dist != b.dist {
			if a.dist < b.dist {
				return -1
			}
			return 1
		}
		return strings.Compare(a.track.Track().Title().String(), b.track.Track().Title().String())
	})

	if len(albumHits) > maxSearchItems {
		albumHits = albumHits[:maxSearchItems]
	}
	if len(trackHits) > maxSearchItems {
		trackHits = trackHits[:maxSearchItems]
	}

	albums := make([]FoundAlbum, 0, len(albumHits))
	for _, h := range albumHits {
		albums = append(albums, FoundAlbum{Album: h.album, Dist: h.dist})
	}
	tracks := make([]FoundTrack, 0, len(trackHits))
	for _, h := range trackHits {
		tracks = append(tracks, FoundTrack{Track: h.track, Dist: h.dist})
	}
	return NewSearchResults(albums, tracks)
}
