package music_test

import (
	"testing"

	"github.com/boreq/eggplant/application/music"
	"github.com/stretchr/testify/require"
)

func TestIfNoTracksAndAlbumsThenReturnForbidden(t *testing.T) {
	l := mockLibrary{}

	h := music.NewBrowseHandler(l)

	cmd := music.Browse{
		Ids: []music.AlbumId{
			"a",
			"b",
		},
		PublicOnly: false,
	}

	_, err := h.Execute(cmd)
	require.ErrorIs(t, err, music.ErrForbidden)
}

func TestIfNoTracksAndAlbumsButThisIsTheRootDoNotReturnForbidden(t *testing.T) {
	l := mockLibrary{}

	h := music.NewBrowseHandler(l)

	cmd := music.Browse{
		Ids:        nil,
		PublicOnly: false,
	}

	_, err := h.Execute(cmd)
	require.NoError(t, err)
}

type mockLibrary struct {
}

func (mockLibrary) Browse(ids []music.AlbumId, publicOnly bool) (music.Album, error) {
	return music.Album{}, nil
}
