package library_test

import (
	"testing"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/stretchr/testify/require"
)

func TestGetRootAlbumAccess(t *testing.T) {
	baseTree := func(rootVis *library.Visibility) rootSpec {
		return rootSpec{
			visibility: rootVis,
			thumb:      "thumbnail",
			albums: []albumSpec{
				{title: "public_album", visibility: visPublic},
				{title: "private_album", visibility: visPrivate},
				{title: "default_album", visibility: visDefault},
			},
			tracks: []string{"song"},
		}
	}
	cases := []struct {
		name               string
		rootVis            *library.Visibility
		access             library.AccessContext
		notAllowedToAccess bool
		expectAlbums       []string
		expectTracks       []string
	}{
		{
			name:         "anonymous_default_root",
			rootVis:      visDefault,
			access:       anonymous,
			expectAlbums: []string{"public_album"},
			expectTracks: []string{"song"},
		},
		{
			name:               "anonymous_private_root",
			rootVis:            visPrivate,
			access:             anonymous,
			notAllowedToAccess: true,
		},
		{
			name:         "anonymous_public_root",
			rootVis:      visPublic,
			access:       anonymous,
			expectAlbums: []string{"public_album", "default_album"},
			expectTracks: []string{"song"},
		},
		{
			name:         "logged_in_default_root",
			rootVis:      visDefault,
			access:       loggedIn,
			expectAlbums: []string{"public_album", "private_album", "default_album"},
			expectTracks: []string{"song"},
		},
		{
			name:         "logged_in_private_root",
			rootVis:      visPrivate,
			access:       loggedIn,
			expectAlbums: []string{"public_album", "private_album", "default_album"},
			expectTracks: []string{"song"},
		},
		{
			name:         "logged_in_public_root",
			rootVis:      visPublic,
			access:       loggedIn,
			expectAlbums: []string{"public_album", "private_album", "default_album"},
			expectTracks: []string{"song"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := buildLib(t, baseTree(tc.rootVis))
			got, albumErr := lib.GetRootAlbum(tc.access)
			_, trackErr := lib.GetTrack(tc.access, trackIdFor(t, "song"))
			_, thumbErr := lib.GetThumbnail(tc.access, thumbnailIdFor(t, "thumbnail"))
			if tc.notAllowedToAccess {
				require.ErrorIs(t, albumErr, library.ErrAlbumNotFound)
				require.ErrorIs(t, trackErr, library.ErrTrackNotFound)
				require.ErrorIs(t, thumbErr, library.ErrThumbnailNotFound)
			} else {
				require.NoError(t, albumErr)
				require.NoError(t, trackErr)
				require.NoError(t, thumbErr)
				require.ElementsMatch(t, tc.expectAlbums, childAlbumTitles(got.Albums()))
				require.ElementsMatch(t, tc.expectTracks, trackListTitles(got.Tracks().Items()))
			}
		})
	}
}

func TestGetAlbumAccess(t *testing.T) {
	baseTree := func(rootVis, parentVis *library.Visibility) rootSpec {
		return rootSpec{
			visibility: rootVis,
			albums: []albumSpec{
				{
					title:      "parent",
					visibility: parentVis,
					thumb:      "parent_thumbnail",
					tracks:     []string{"parent_song"},
					albums: []albumSpec{
						{title: "public_child", visibility: visPublic, tracks: []string{"public_child_song"}},
						{title: "private_child", visibility: visPrivate, tracks: []string{"private_child_song"}},
					},
				},
			},
		}
	}
	cases := []struct {
		name         string
		rootVis      *library.Visibility
		parentVis    *library.Visibility
		access       library.AccessContext
		expectErr    error
		expectAlbums []string
	}{
		{
			name:         "anonymous_public_root_public_parent",
			rootVis:      visPublic,
			parentVis:    visPublic,
			access:       anonymous,
			expectAlbums: []string{"public_child"},
		},
		{
			name:      "anonymous_public_root_private_parent",
			rootVis:   visPublic,
			parentVis: visPrivate,
			access:    anonymous,
			expectErr: library.ErrAlbumNotFound,
		},
		{
			name:         "anonymous_public_root_default_parent",
			rootVis:      visPublic,
			parentVis:    visDefault,
			access:       anonymous,
			expectAlbums: []string{"public_child"},
		},
		{
			name:         "anonymous_private_root_public_parent",
			rootVis:      visPrivate,
			parentVis:    visPublic,
			access:       anonymous,
			expectAlbums: []string{"public_child"},
		},
		{
			name:      "anonymous_private_root_private_parent",
			rootVis:   visPrivate,
			parentVis: visPrivate,
			access:    anonymous,
			expectErr: library.ErrAlbumNotFound,
		},
		{
			name:      "anonymous_private_root_default_parent",
			rootVis:   visPrivate,
			parentVis: visDefault,
			access:    anonymous,
			expectErr: library.ErrAlbumNotFound,
		},
		{
			name:         "anonymous_default_root_public_parent",
			rootVis:      visDefault,
			parentVis:    visPublic,
			access:       anonymous,
			expectAlbums: []string{"public_child"},
		},
		{
			name:      "anonymous_default_root_private_parent",
			rootVis:   visDefault,
			parentVis: visPrivate,
			access:    anonymous,
			expectErr: library.ErrAlbumNotFound,
		},
		{
			name:      "anonymous_default_root_default_parent",
			rootVis:   visDefault,
			parentVis: visDefault,
			access:    anonymous,
			expectErr: library.ErrAlbumNotFound,
		},
		{
			name:         "logged_in_public_root_public_parent",
			rootVis:      visPublic,
			parentVis:    visPublic,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_public_root_private_parent",
			rootVis:      visPublic,
			parentVis:    visPrivate,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_public_root_default_parent",
			rootVis:      visPublic,
			parentVis:    visDefault,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_private_root_public_parent",
			rootVis:      visPrivate,
			parentVis:    visPublic,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_private_root_private_parent",
			rootVis:      visPrivate,
			parentVis:    visPrivate,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_private_root_default_parent",
			rootVis:      visPrivate,
			parentVis:    visDefault,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_default_root_public_parent",
			rootVis:      visDefault,
			parentVis:    visPublic,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_default_root_private_parent",
			rootVis:      visDefault,
			parentVis:    visPrivate,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
		{
			name:         "logged_in_default_root_default_parent",
			rootVis:      visDefault,
			parentVis:    visDefault,
			access:       loggedIn,
			expectAlbums: []string{"public_child", "private_child"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := buildLib(t, baseTree(tc.rootVis, tc.parentVis))
			got, albumErr := lib.GetAlbum(tc.access, albumIdFor(t, "parent"))
			_, trackErr := lib.GetTrack(tc.access, trackIdFor(t, "parent_song"))
			_, thumbErr := lib.GetThumbnail(tc.access, thumbnailIdFor(t, "parent_thumbnail"))
			if tc.expectErr != nil {
				require.ErrorIs(t, albumErr, tc.expectErr)
				require.ErrorIs(t, trackErr, library.ErrTrackNotFound)
				require.ErrorIs(t, thumbErr, library.ErrThumbnailNotFound)
				return
			}
			require.NoError(t, albumErr)
			require.NoError(t, trackErr)
			require.NoError(t, thumbErr)
			require.Equal(t, "parent", got.Title().String())
			require.ElementsMatch(t, tc.expectAlbums, childAlbumTitles(got.Albums()))
			require.ElementsMatch(t, []string{"parent_song"}, trackListTitles(got.Tracks().Items()))
		})
	}
}

func TestGetGrandchildAlbumAccess(t *testing.T) {
	baseTree := func(parentVis, grandchildVis *library.Visibility) rootSpec {
		return rootSpec{
			visibility: visDefault,
			albums: []albumSpec{
				{
					title:      "parent",
					visibility: parentVis,
					tracks:     []string{"parent_song"},
					albums: []albumSpec{
						{
							title:      "grandchild",
							visibility: grandchildVis,
							thumb:      "grandchild_thumbnail",
							tracks:     []string{"grandchild_song"},
							albums: []albumSpec{
								{title: "public_child", visibility: visPublic, tracks: []string{"public_child_song"}},
								{title: "private_child", visibility: visPrivate, tracks: []string{"private_child_song"}},
							},
						},
					},
				},
			},
		}
	}
	cases := []struct {
		name          string
		parentVis     *library.Visibility
		grandchildVis *library.Visibility
		access        library.AccessContext
		expectErr     error
		expectAlbums  []string
	}{
		{
			name:          "anonymous_public_parent_public_grandchild",
			parentVis:     visPublic,
			grandchildVis: visPublic,
			access:        anonymous,
			expectAlbums:  []string{"public_child"},
		},
		{
			name:          "anonymous_public_parent_private_grandchild",
			parentVis:     visPublic,
			grandchildVis: visPrivate,
			access:        anonymous,
			expectErr:     library.ErrAlbumNotFound,
		},
		{
			name:          "anonymous_public_parent_default_grandchild",
			parentVis:     visPublic,
			grandchildVis: visDefault,
			access:        anonymous,
			expectAlbums:  []string{"public_child"},
		},
		{
			name:          "anonymous_private_parent_public_grandchild",
			parentVis:     visPrivate,
			grandchildVis: visPublic,
			access:        anonymous,
			expectAlbums:  []string{"public_child"},
		},
		{
			name:          "anonymous_private_parent_private_grandchild",
			parentVis:     visPrivate,
			grandchildVis: visPrivate,
			access:        anonymous,
			expectErr:     library.ErrAlbumNotFound,
		},
		{
			name:          "anonymous_private_parent_default_grandchild",
			parentVis:     visPrivate,
			grandchildVis: visDefault,
			access:        anonymous,
			expectErr:     library.ErrAlbumNotFound,
		},
		{
			name:          "anonymous_default_parent_public_grandchild",
			parentVis:     visDefault,
			grandchildVis: visPublic,
			access:        anonymous,
			expectAlbums:  []string{"public_child"},
		},
		{
			name:          "anonymous_default_parent_private_grandchild",
			parentVis:     visDefault,
			grandchildVis: visPrivate,
			access:        anonymous,
			expectErr:     library.ErrAlbumNotFound,
		},
		{
			name:          "anonymous_default_parent_default_grandchild",
			parentVis:     visDefault,
			grandchildVis: visDefault,
			access:        anonymous,
			expectErr:     library.ErrAlbumNotFound,
		},
		{
			name:          "logged_in_public_parent_public_grandchild",
			parentVis:     visPublic,
			grandchildVis: visPublic,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_public_parent_private_grandchild",
			parentVis:     visPublic,
			grandchildVis: visPrivate,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_public_parent_default_grandchild",
			parentVis:     visPublic,
			grandchildVis: visDefault,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_private_parent_public_grandchild",
			parentVis:     visPrivate,
			grandchildVis: visPublic,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_private_parent_private_grandchild",
			parentVis:     visPrivate,
			grandchildVis: visPrivate,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_private_parent_default_grandchild",
			parentVis:     visPrivate,
			grandchildVis: visDefault,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_default_parent_public_grandchild",
			parentVis:     visDefault,
			grandchildVis: visPublic,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_default_parent_private_grandchild",
			parentVis:     visDefault,
			grandchildVis: visPrivate,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
		{
			name:          "logged_in_default_parent_default_grandchild",
			parentVis:     visDefault,
			grandchildVis: visDefault,
			access:        loggedIn,
			expectAlbums:  []string{"public_child", "private_child"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := buildLib(t, baseTree(tc.parentVis, tc.grandchildVis))
			got, albumErr := lib.GetAlbum(tc.access, albumIdFor(t, "grandchild"))
			_, trackErr := lib.GetTrack(tc.access, trackIdFor(t, "grandchild_song"))
			_, thumbErr := lib.GetThumbnail(tc.access, thumbnailIdFor(t, "grandchild_thumbnail"))
			if tc.expectErr != nil {
				require.ErrorIs(t, albumErr, tc.expectErr)
				require.ErrorIs(t, trackErr, library.ErrTrackNotFound)
				require.ErrorIs(t, thumbErr, library.ErrThumbnailNotFound)
				return
			}
			require.NoError(t, albumErr)
			require.NoError(t, trackErr)
			require.NoError(t, thumbErr)
			require.Equal(t, "grandchild", got.Title().String())
			require.ElementsMatch(t, tc.expectAlbums, childAlbumTitles(got.Albums()))
			require.ElementsMatch(t, []string{"grandchild_song"}, trackListTitles(got.Tracks().Items()))
		})
	}
}

func TestGetAlbumNotFound(t *testing.T) {
	lib := buildLib(t, rootSpec{
		visibility: visPublic,
		albums: []albumSpec{
			{title: "present", visibility: visPublic},
		},
	})
	_, err := lib.GetAlbum(loggedIn, albumIdFor(t, "nonexistent"))
	require.ErrorIs(t, err, library.ErrAlbumNotFound)
}

func TestGetAlbumWithNoVisibleContent(t *testing.T) {
	lib := buildLib(t, rootSpec{
		visibility: visPublic,
		albums: []albumSpec{
			{
				title:      "parent",
				visibility: visPublic,
				albums: []albumSpec{
					{title: "private_child", visibility: visPrivate, tracks: []string{"private_child_song"}},
				},
			},
		},
	})
	_, err := lib.GetAlbum(anonymous, albumIdFor(t, "parent"))
	require.ErrorIs(t, err, library.ErrAlbumNotFound)
}

func TestGetTrackNotFound(t *testing.T) {
	lib := buildLib(t, rootSpec{
		visibility: visPublic,
		albums: []albumSpec{
			{title: "album", visibility: visPublic, tracks: []string{"present"}},
		},
	})
	_, err := lib.GetTrack(loggedIn, trackIdFor(t, "nope"))
	require.ErrorIs(t, err, library.ErrTrackNotFound)
}

func TestGetThumbnailNotFound(t *testing.T) {
	lib := buildLib(t, rootSpec{
		visibility: visPublic,
		albums: []albumSpec{
			{title: "album", visibility: visPublic},
		},
	})
	_, err := lib.GetThumbnail(loggedIn, thumbnailIdFor(t, "nope.jpg"))
	require.ErrorIs(t, err, library.ErrThumbnailNotFound)
}

var (
	anonymous library.AccessContext = library.NewAnonymousAccessContext()
	loggedIn  library.AccessContext = library.NewLoggedInAccessContext()

	visPublic  = vis(true)
	visPrivate = vis(false)
	visDefault *library.Visibility
)

type albumSpec struct {
	title      string
	visibility *library.Visibility
	tracks     []string
	thumb      string
	albums     []albumSpec
}

type rootSpec struct {
	visibility *library.Visibility
	tracks     []string
	thumb      string
	albums     []albumSpec
}

func vis(v bool) *library.Visibility {
	out := library.NewVisibility(v)
	return &out
}

func buildLib(t *testing.T, spec rootSpec) *library.Library {
	t.Helper()

	var topAlbums []library.Album
	for _, a := range spec.albums {
		topAlbums = append(topAlbums, buildAlbum(t, a))
	}

	var rootTracks []music.Track
	for _, title := range spec.tracks {
		rootTracks = append(rootTracks, mkTrack(t, title))
	}

	var rootThumb *music.Thumbnail
	if spec.thumb != "" {
		th := mkThumb(t, spec.thumb)
		rootThumb = &th
	}

	r, err := library.NewRootAlbum(rootThumb, spec.visibility, topAlbums, rootTracks)
	require.NoError(t, err)
	return library.NewLibrary(r)
}

func buildAlbum(t *testing.T, spec albumSpec) library.Album {
	t.Helper()

	id := albumIdFor(t, spec.title)
	title, err := music.NewAlbumTitle(spec.title)
	require.NoError(t, err)

	var subAlbums []library.Album
	for _, a := range spec.albums {
		subAlbums = append(subAlbums, buildAlbum(t, a))
	}

	var tracks []music.Track
	for _, trackTitle := range spec.tracks {
		tracks = append(tracks, mkTrack(t, trackTitle))
	}

	if len(subAlbums) == 0 && len(tracks) == 0 {
		tracks = []music.Track{mkTrack(t, "_placeholder_"+spec.title)}
	}

	var thumb *music.Thumbnail
	if spec.thumb != "" {
		th := mkThumb(t, spec.thumb)
		thumb = &th
	}

	a, err := library.NewAlbum(id, title, thumb, spec.visibility, subAlbums, tracks)
	require.NoError(t, err)
	return a
}

func albumIdFor(t *testing.T, title string) music.AlbumId {
	t.Helper()
	at, err := music.NewAlbumTitle(title)
	require.NoError(t, err)
	id, err := music.NewAlbumId(nil, at)
	require.NoError(t, err)
	return id
}

func trackIdFor(t *testing.T, title string) music.TrackId {
	t.Helper()
	tt, err := music.NewTrackTitle(title)
	require.NoError(t, err)
	id, err := music.NewTrackId(nil, tt)
	require.NoError(t, err)
	return id
}

func thumbnailIdFor(t *testing.T, filename string) music.ThumbnailId {
	t.Helper()
	name, err := music.NewFileName(filename)
	require.NoError(t, err)
	id, err := music.NewThumbnailId(nil, name)
	require.NoError(t, err)
	return id
}

func mkTrack(t *testing.T, title string) music.Track {
	t.Helper()
	tt, err := music.NewTrackTitle(title)
	require.NoError(t, err)
	id := trackIdFor(t, title)
	p, err := music.NewFilePath("/test/" + title)
	require.NoError(t, err)
	fileId, err := music.NewFileId(p)
	require.NoError(t, err)
	return music.NewTrack(id, fileId, tt)
}

func mkThumb(t *testing.T, filename string) music.Thumbnail {
	t.Helper()
	id := thumbnailIdFor(t, filename)
	p, err := music.NewFilePath("/test/thumb/" + filename)
	require.NoError(t, err)
	fileId, err := music.NewFileId(p)
	require.NoError(t, err)
	return music.NewThumbnail(id, fileId)
}

func trackTitles(tracks []library.TrackWithAlbum) []string {
	titles := make([]string, 0, len(tracks))
	for _, t := range tracks {
		titles = append(titles, t.Track().Title().String())
	}
	return titles
}

func childAlbumTitles(albums []music.PartialAlbum) []string {
	titles := make([]string, 0, len(albums))
	for _, a := range albums {
		titles = append(titles, a.Title().String())
	}
	return titles
}

func trackListTitles(tracks []music.Track) []string {
	titles := make([]string, 0, len(tracks))
	for _, t := range tracks {
		titles = append(titles, t.Title().String())
	}
	return titles
}
