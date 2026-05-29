package music_test

import (
	"testing"

	"github.com/boreq/eggplant/domain/music"
	"github.com/stretchr/testify/require"
)

func TestNewAlbum(t *testing.T) {
	selfTitle, err := music.NewAlbumTitle("self")
	require.NoError(t, err)
	selfId, err := music.NewAlbumId(nil, selfTitle)
	require.NoError(t, err)

	otherTitle, err := music.NewAlbumTitle("other")
	require.NoError(t, err)
	otherId, err := music.NewAlbumId(nil, otherTitle)
	require.NoError(t, err)

	tracks := []music.Track{validTrack(t)}

	testCases := []struct {
		name    string
		parents []music.PartialAlbum
		albums  []music.PartialAlbum
		tracks  []music.Track
		errMsg  string
	}{
		{
			name:    "rejects_parents_containing_self",
			parents: []music.PartialAlbum{music.NewPartialAlbum(selfId, selfTitle, nil)},
			tracks:  tracks,
			errMsg:  "parents must not contain the album itself",
		},
		{
			name:   "rejects_children_containing_self",
			albums: []music.PartialAlbum{music.NewPartialAlbum(selfId, selfTitle, nil)},
			errMsg: "child albums must not contain the album itself",
		},
		{
			name:   "rejects_album_with_no_children_and_no_tracks",
			errMsg: "album must have at least one child album or track",
		},
		{
			name:   "accepts_top_level_album_with_no_parents",
			tracks: tracks,
		},
		{
			name:    "accepts_actual_parents",
			parents: []music.PartialAlbum{music.NewPartialAlbum(otherId, otherTitle, nil)},
			tracks:  tracks,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := music.NewAlbum(selfId, selfTitle, nil, tc.parents, tc.albums, tc.tracks)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.errMsg)
			}
		})
	}
}

func validTrack(t *testing.T) music.Track {
	t.Helper()
	title, err := music.NewTrackTitle("track")
	require.NoError(t, err)
	return music.NewTrack(music.TrackId{}, music.FileId{}, title)
}
