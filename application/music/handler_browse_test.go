package music_test

import (
	"testing"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
)

type mockLibrary struct {
}

func (mockLibrary) Browse(ids []music.AlbumId, publicOnly bool) (music.Album, error) {
	return music.Album{}, nil
}

func TestIfNoTracksAndAlbumsThenReturnForbidden(t *testing.T) {
	l := mockLibrary{}

	h := music.NewBrowseHandler(l)

	cmd := music.Browse{
		Ids:        nil,
		PublicOnly: false,
	}

	_, err := h.Execute(cmd)
	require.True(t, errors.Is(err, music.ErrForbidden))
}
