package domain_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/domain"
	"github.com/stretchr/testify/require"
)

func TestNewAlbum(t *testing.T) {
	selfTitle, err := domain.NewAlbumTitle("self")
	require.NoError(t, err)
	selfId, err := domain.NewAlbumId(nil, selfTitle)
	require.NoError(t, err)

	otherTitle, err := domain.NewAlbumTitle("other")
	require.NoError(t, err)
	otherId, err := domain.NewAlbumId(nil, otherTitle)
	require.NoError(t, err)

	tracks := []domain.Track{validTrack(t)}

	testCases := []struct {
		name    string
		parents []domain.ParentAlbum
		albums  []domain.ChildAlbum
		tracks  []domain.Track
		errMsg  string
	}{
		{
			name:    "rejects_parents_containing_self",
			parents: []domain.ParentAlbum{domain.NewParentAlbum(selfId, selfTitle)},
			tracks:  tracks,
			errMsg:  "parents must not contain the album itself",
		},
		{
			name:   "rejects_children_containing_self",
			albums: []domain.ChildAlbum{domain.NewChildAlbum(selfId, selfTitle, nil)},
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
			parents: []domain.ParentAlbum{domain.NewParentAlbum(otherId, otherTitle)},
			tracks:  tracks,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.NewAlbum(selfId, selfTitle, nil, tc.parents, tc.albums, tc.tracks)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.errMsg)
			}
		})
	}
}

func validTrack(t *testing.T) domain.Track {
	t.Helper()
	title, err := domain.NewTrackTitle("track")
	require.NoError(t, err)
	dur, err := domain.NewTrackDuration(time.Second)
	require.NoError(t, err)
	return domain.NewTrack(domain.TrackId{}, domain.FileId{}, title, dur)
}
