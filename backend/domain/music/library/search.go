package library

import (
	"slices"
	"strings"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

const maxSearchItems = 20

func (l *Library) Search(accessCtx AccessContext, query string) (SearchResults, error) {
	builder := newSearchBuilder()

	if accessCtx.CanSee(l.getRootVisibility()) {
		for _, track := range l.root.tracks {
			if !containsStringCaseInsensitive(track.Title().String(), query) {
				continue
			}
			if err := builder.addTrack(NewRootTrackWithAlbum(track), 0); err != nil {
				return SearchResults{}, errors.Wrap(err, "could not add a track")
			}
		}
	}

	if err := searchAlbums(builder, l.root.albums, nil, l.getRootVisibilityToPropagateToChildren(), query, accessCtx); err != nil {
		return SearchResults{}, errors.Wrap(err, "could not search the albums")
	}

	return builder.build(), nil
}

func searchAlbums(b *searchBuilder, albums []Album, parentInheritedDist *int, parentVis Visibility, query string, ctx AccessContext) error {
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
				if err := b.addAlbum(album, *ownDist); err != nil {
					return errors.Wrap(err, "could not add an album")
				}
			}

			childDist := incDist(ownDist)
			for _, track := range a.tracks {
				dist := childDist
				if containsStringCaseInsensitive(track.Title().String(), query) {
					dist = new(0)
				}
				if dist == nil {
					continue
				}
				if err := b.addTrack(NewTrackWithAlbum(track, album), *dist); err != nil {
					return errors.Wrap(err, "could not add a track")
				}
			}
		}

		if err := searchAlbums(b, a.albums, incDist(ownDist), effective, query, ctx); err != nil {
			return err
		}
	}
	return nil
}

type FoundAlbum struct {
	album music.PartialAlbum
	dist  int
}

func NewFoundAlbum(album music.PartialAlbum, dist int) (FoundAlbum, error) {
	if dist < 0 {
		return FoundAlbum{}, errors.New("negative distance")
	}
	return FoundAlbum{album: album, dist: dist}, nil
}

func (a FoundAlbum) Album() music.PartialAlbum {
	return a.album
}

func (a FoundAlbum) Dist() int {
	return a.dist
}

type FoundTrack struct {
	track TrackWithAlbum
	dist  int
}

func NewFoundTrack(track TrackWithAlbum, dist int) (FoundTrack, error) {
	if dist < 0 {
		return FoundTrack{}, errors.New("negative distance")
	}
	return FoundTrack{track: track, dist: dist}, nil
}

func (t FoundTrack) Track() TrackWithAlbum {
	return t.track
}

func (t FoundTrack) Dist() int {
	return t.dist
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

func MergeResults(results ...SearchResults) (SearchResults, error) {
	b := newSearchBuilder()
	for _, r := range results {
		for _, h := range r.albums {
			if err := b.addAlbum(h.album, h.dist); err != nil {
				return SearchResults{}, errors.Wrap(err, "could not add an album")
			}
		}
		for _, h := range r.tracks {
			if err := b.addTrack(h.track, h.dist); err != nil {
				return SearchResults{}, errors.Wrap(err, "could not add a track")
			}
		}
	}
	return b.build(), nil
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

type searchBuilder struct {
	albums map[string]FoundAlbum
	tracks map[string]FoundTrack
}

func newSearchBuilder() *searchBuilder {
	return &searchBuilder{
		albums: map[string]FoundAlbum{},
		tracks: map[string]FoundTrack{},
	}
}

func (b *searchBuilder) addAlbum(a music.PartialAlbum, dist int) error {
	key := instanceKey(a.RemoteInstanceId()) + a.Id().String()
	if existing, ok := b.albums[key]; ok && existing.dist <= dist {
		return nil
	}
	found, err := NewFoundAlbum(a, dist)
	if err != nil {
		return errors.Wrap(err, "could not create a found album")
	}
	b.albums[key] = found
	return nil
}

func (b *searchBuilder) addTrack(t TrackWithAlbum, dist int) error {
	key := instanceKey(t.Track().RemoteInstanceId()) + t.Track().Id().String()
	if existing, ok := b.tracks[key]; ok && existing.dist <= dist {
		return nil
	}
	found, err := NewFoundTrack(t, dist)
	if err != nil {
		return errors.Wrap(err, "could not create a found track")
	}
	b.tracks[key] = found
	return nil
}

func instanceKey(id *remote.RemoteInstanceID) string {
	if id == nil {
		return ""
	}
	return id.String() + "/"
}

func (b *searchBuilder) build() SearchResults {
	albums := make([]FoundAlbum, 0, len(b.albums))
	for _, h := range b.albums {
		albums = append(albums, h)
	}
	slices.SortStableFunc(albums, func(a, b FoundAlbum) int {
		if a.dist != b.dist {
			if a.dist < b.dist {
				return -1
			}
			return 1
		}
		return strings.Compare(a.album.Title().String(), b.album.Title().String())
	})

	tracks := make([]FoundTrack, 0, len(b.tracks))
	for _, h := range b.tracks {
		tracks = append(tracks, h)
	}
	slices.SortStableFunc(tracks, func(a, b FoundTrack) int {
		if a.dist != b.dist {
			if a.dist < b.dist {
				return -1
			}
			return 1
		}
		return strings.Compare(a.track.Track().Title().String(), b.track.Track().Title().String())
	})

	if len(albums) > maxSearchItems {
		albums = albums[:maxSearchItems]
	}
	if len(tracks) > maxSearchItems {
		tracks = tracks[:maxSearchItems]
	}

	return NewSearchResults(albums, tracks)
}
