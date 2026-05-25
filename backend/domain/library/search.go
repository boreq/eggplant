package library

import (
	"sort"
	"strings"

	"github.com/boreq/eggplant/domain"
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
			ownDist = ptr(0)
		} else {
			ownDist = parentInheritedDist
		}

		if ctx.CanSee(effective) {
			album := domain.NewPartialAlbum(a.id, a.title, a.thumbnail)

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

type SearchResults struct {
	albums []domain.PartialAlbum
	tracks []TrackWithAlbum
}

func NewSearchResults(albums []domain.PartialAlbum, tracks []TrackWithAlbum) SearchResults {
	return SearchResults{albums: albums, tracks: tracks}
}

func (s SearchResults) Albums() []domain.PartialAlbum {
	return s.albums
}

func (s SearchResults) Tracks() []TrackWithAlbum {
	return s.tracks
}

type TrackWithAlbum struct {
	track domain.Track
	album *domain.PartialAlbum
}

func NewRootTrackWithAlbum(track domain.Track) TrackWithAlbum {
	return TrackWithAlbum{track: track, album: nil}
}

func NewTrackWithAlbum(track domain.Track, album domain.PartialAlbum) TrackWithAlbum {
	return TrackWithAlbum{track: track, album: &album}
}

func (t TrackWithAlbum) Track() domain.Track {
	return t.track
}

func (t TrackWithAlbum) Album() *domain.PartialAlbum {
	return t.album
}

func incDist(d *int) *int {
	if d == nil {
		return nil
	}
	return ptr(*d + 1)
}

func ptr(v int) *int {
	return &v
}

func containsStringCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

type albumHit struct {
	album domain.PartialAlbum
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

func (b *searchBuilder) addAlbum(a domain.PartialAlbum, dist int) {
	key := a.Id().String()
	if existing, ok := b.albums[key]; ok && existing.dist <= dist {
		return
	}
	b.albums[key] = albumHit{album: a, dist: dist}
}

func (b *searchBuilder) addTrack(t TrackWithAlbum, dist int) {
	key := t.Track().Id().String()
	if existing, ok := b.tracks[key]; ok && existing.dist <= dist {
		return
	}
	b.tracks[key] = trackHit{track: t, dist: dist}
}

func (b *searchBuilder) build() SearchResults {
	albumHits := make([]albumHit, 0, len(b.albums))
	for _, h := range b.albums {
		albumHits = append(albumHits, h)
	}
	sort.SliceStable(albumHits, func(i, j int) bool {
		if albumHits[i].dist != albumHits[j].dist {
			return albumHits[i].dist < albumHits[j].dist
		}
		return albumHits[i].album.Title().String() < albumHits[j].album.Title().String()
	})

	trackHits := make([]trackHit, 0, len(b.tracks))
	for _, h := range b.tracks {
		trackHits = append(trackHits, h)
	}
	sort.SliceStable(trackHits, func(i, j int) bool {
		if trackHits[i].dist != trackHits[j].dist {
			return trackHits[i].dist < trackHits[j].dist
		}
		return trackHits[i].track.Track().Title().String() < trackHits[j].track.Track().Title().String()
	})

	if len(albumHits) > maxSearchItems {
		albumHits = albumHits[:maxSearchItems]
	}
	if len(trackHits) > maxSearchItems {
		trackHits = trackHits[:maxSearchItems]
	}

	albums := make([]domain.PartialAlbum, 0, len(albumHits))
	for _, h := range albumHits {
		albums = append(albums, h.album)
	}
	tracks := make([]TrackWithAlbum, 0, len(trackHits))
	for _, h := range trackHits {
		tracks = append(tracks, h.track)
	}
	return NewSearchResults(albums, tracks)
}
