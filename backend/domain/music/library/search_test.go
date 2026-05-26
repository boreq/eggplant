package library_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/boreq/eggplant/domain/music/library"
	"github.com/stretchr/testify/require"
)

func TestSearchAccess(t *testing.T) {
	baseTree := func(rootVis *library.Visibility) rootSpec {
		return rootSpec{
			visibility: rootVis,
			tracks:     []string{"root_song"},
			albums: []albumSpec{
				{title: "public_album", visibility: visPublic, tracks: []string{"public_song"}},
				{title: "private_album", visibility: visPrivate, tracks: []string{"private_song"}},
				{
					title:      "private_parent",
					visibility: visPrivate,
					tracks:     []string{"hidden_song"},
					albums: []albumSpec{
						{title: "public_child", visibility: visPublic, tracks: []string{"deep_song"}},
					},
				},
			},
		}
	}
	cases := []struct {
		name         string
		rootVis      *library.Visibility
		query        string
		access       library.AccessContext
		expectTitles []string
	}{
		{
			name:         "anonymous_sees_public_and_override_hits",
			rootVis:      visPublic,
			query:        "song",
			access:       anonymous,
			expectTitles: []string{"root_song", "public_song", "deep_song"},
		},
		{
			name:         "logged_in_sees_all_hits",
			rootVis:      visPublic,
			query:        "song",
			access:       loggedIn,
			expectTitles: []string{"root_song", "public_song", "private_song", "hidden_song", "deep_song"},
		},
		{
			name:         "anonymous_blocked_from_root_track_when_root_private",
			rootVis:      visPrivate,
			query:        "song",
			access:       anonymous,
			expectTitles: []string{"public_song", "deep_song"},
		},
		{
			name:         "anonymous_sees_root_track_when_root_default",
			rootVis:      visDefault,
			query:        "song",
			access:       anonymous,
			expectTitles: []string{"root_song", "public_song", "deep_song"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := buildLib(t, baseTree(tc.rootVis))
			result, err := lib.Search(tc.access, tc.query)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.expectTitles, trackTitles(result.Tracks()))
		})
	}
}

func TestSearchContent(t *testing.T) {
	t.Run("album_hit_carries_path_and_title", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "parent",
					visibility: visPublic,
					albums: []albumSpec{
						{title: "findme", visibility: visPublic},
					},
				},
			},
		})
		result, err := lib.Search(anonymous, "findme")
		require.NoError(t, err)
		require.Len(t, result.Albums(), 1)
		require.Equal(t, "findme", result.Albums()[0].Title().String())
		require.Equal(t, albumIdFor(t, "findme"), result.Albums()[0].Id())
	})

	t.Run("track_hit_carries_containing_album_ref", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{title: "parent", visibility: visPublic, tracks: []string{"findme"}},
			},
		})
		result, err := lib.Search(anonymous, "findme")
		require.NoError(t, err)
		require.Len(t, result.Tracks(), 1)
		require.NotNil(t, result.Tracks()[0].Album())
		require.Equal(t, "parent", result.Tracks()[0].Album().Title().String())
		require.Equal(t, albumIdFor(t, "parent"), result.Tracks()[0].Album().Id())
	})

	t.Run("root_track_hit_has_no_album_ref", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			tracks:     []string{"findme"},
			albums: []albumSpec{
				{title: "placeholder", visibility: visPublic},
			},
		})
		result, err := lib.Search(anonymous, "findme")
		require.NoError(t, err)
		require.Len(t, result.Tracks(), 1)
		require.Nil(t, result.Tracks()[0].Album())
	})

	t.Run("no_matches_returns_empty_result", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			tracks:     []string{"beta"},
			albums: []albumSpec{
				{title: "alpha", visibility: visPublic},
			},
		})
		result, err := lib.Search(loggedIn, "zzz")
		require.NoError(t, err)
		require.Empty(t, result.Albums())
		require.Empty(t, result.Tracks())
	})

	t.Run("parent_album_match_pulls_in_tracks_and_subalbums", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "band_a",
					visibility: visPublic,
					tracks:     []string{"track_one", "track_two"},
					albums: []albumSpec{
						{title: "sub_album", visibility: visPublic, tracks: []string{"sub_track"}},
					},
				},
				{title: "unrelated", visibility: visPublic, tracks: []string{"other"}},
			},
		})
		result, err := lib.Search(anonymous, "band_a")
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"band_a", "sub_album"}, childAlbumTitles(result.Albums()))
		require.ElementsMatch(t, []string{"track_one", "track_two", "sub_track"}, trackTitles(result.Tracks()))
	})
}

func TestSearchDistance(t *testing.T) {
	t.Run("grandparent_match_includes_descendants", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "needle",
					visibility: visPublic,
					albums: []albumSpec{
						{
							title:      "middle",
							visibility: visPublic,
							tracks:     []string{"leaf_track"},
							albums: []albumSpec{
								{title: "deep", visibility: visPublic, tracks: []string{"deep_track"}},
							},
						},
					},
				},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		require.Equal(t, []string{"needle", "middle", "deep"}, childAlbumTitles(result.Albums()))
		require.Equal(t, []string{"leaf_track", "deep_track"}, trackTitles(result.Tracks()))
	})

	t.Run("own_match_outranks_ancestor_match", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "needle_parent",
					visibility: visPublic,
					tracks:     []string{"inherited_one", "inherited_two", "needle_own"},
				},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		require.Equal(t, "needle_own", result.Tracks()[0].Track().Title().String())
	})

	t.Run("album_distances_ordered", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "needle_grandparent",
					visibility: visPublic,
					albums: []albumSpec{
						{
							title:      "parent",
							visibility: visPublic,
							albums: []albumSpec{
								{title: "leaf_a", visibility: visPublic, tracks: []string{"track_a"}},
							},
						},
					},
				},
				{title: "needle_own", visibility: visPublic, tracks: []string{"unrelated"}},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		titles := childAlbumTitles(result.Albums())
		require.Equal(t, []string{"needle_grandparent", "needle_own", "parent", "leaf_a"}, titles)
	})

	t.Run("dedup_keeps_each_id_once", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "foo",
					visibility: visPublic,
					tracks:     []string{"foo_song"},
				},
			},
		})
		result, err := lib.Search(anonymous, "foo")
		require.NoError(t, err)
		require.Len(t, result.Albums(), 1)
		require.Len(t, result.Tracks(), 1)
	})
}

func TestSearchCaseInsensitive(t *testing.T) {
	lib := buildLib(t, rootSpec{
		visibility: visPublic,
		albums: []albumSpec{
			{title: "MixedCase", visibility: visPublic, tracks: []string{"Inner"}},
		},
	})
	result, err := lib.Search(anonymous, "mIxEdCaSe")
	require.NoError(t, err)
	require.Equal(t, []string{"MixedCase"}, childAlbumTitles(result.Albums()))
}

func TestSearchSubstringMatch(t *testing.T) {
	lib := buildLib(t, rootSpec{
		visibility: visPublic,
		albums: []albumSpec{
			{title: "long_compound_album_title", visibility: visPublic, tracks: []string{"unrelated_song"}},
		},
	})
	result, err := lib.Search(anonymous, "compound_album")
	require.NoError(t, err)
	require.Len(t, result.Albums(), 1)
	require.True(t, strings.Contains(result.Albums()[0].Title().String(), "compound_album"))
}

func TestSearchCap(t *testing.T) {
	t.Run("more_than_max_items_truncates_after_sort", func(t *testing.T) {
		var subAlbums []albumSpec
		for i := 0; i < 25; i++ {
			subAlbums = append(subAlbums, albumSpec{
				title:      fmt.Sprintf("child_%02d", i),
				visibility: visPublic,
				tracks:     []string{fmt.Sprintf("track_%02d", i)},
			})
		}
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "needle_parent",
					visibility: visPublic,
					albums:     subAlbums,
				},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		require.Len(t, result.Albums(), 20)
		require.Len(t, result.Tracks(), 20)
		require.Equal(t, "needle_parent", result.Albums()[0].Title().String())
	})

	t.Run("own_matches_not_dropped_by_cap", func(t *testing.T) {
		var subAlbums []albumSpec
		for i := 0; i < 25; i++ {
			subAlbums = append(subAlbums, albumSpec{
				title:      fmt.Sprintf("child_%02d", i),
				visibility: visPublic,
				tracks:     []string{fmt.Sprintf("track_%02d", i)},
			})
		}
		subAlbums = append(subAlbums, albumSpec{
			title:      "needle_self",
			visibility: visPublic,
			tracks:     []string{"unrelated"},
		})
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{title: "needle_parent", visibility: visPublic, albums: subAlbums},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		titles := childAlbumTitles(result.Albums())
		require.Contains(t, titles, "needle_self")
		require.Equal(t, "needle_parent", titles[0])
	})
}

func TestSearchVisibilityBlocksAncestorPropagation(t *testing.T) {
	t.Run("private_subalbum_not_pulled_in_by_public_parent_match", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			albums: []albumSpec{
				{
					title:      "needle",
					visibility: visPublic,
					tracks:     []string{"visible_track"},
					albums: []albumSpec{
						{title: "private_child", visibility: visPrivate, tracks: []string{"hidden_track"}},
					},
				},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		require.NotContains(t, childAlbumTitles(result.Albums()), "private_child")
		require.NotContains(t, trackTitles(result.Tracks()), "hidden_track")
		require.Contains(t, trackTitles(result.Tracks()), "visible_track")
	})
}

func TestSearchRootBehavior(t *testing.T) {
	t.Run("root_tracks_only_via_own_match", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			tracks:     []string{"needle_root_song", "other_root_song"},
			albums: []albumSpec{
				{title: "needle_album", visibility: visPublic, tracks: []string{"album_track"}},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		titles := trackTitles(result.Tracks())
		require.Contains(t, titles, "needle_root_song")
		require.NotContains(t, titles, "other_root_song")
	})

	t.Run("root_track_album_ref_remains_nil", func(t *testing.T) {
		lib := buildLib(t, rootSpec{
			visibility: visPublic,
			tracks:     []string{"needle_song"},
			albums: []albumSpec{
				{title: "placeholder", visibility: visPublic},
			},
		})
		result, err := lib.Search(anonymous, "needle")
		require.NoError(t, err)
		for _, twa := range result.Tracks() {
			if twa.Track().Title().String() == "needle_song" {
				require.Nil(t, twa.Album())
				return
			}
		}
		t.Fatal("needle_song not in results")
	})
}
