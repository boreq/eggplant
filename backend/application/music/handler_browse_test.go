package music_test

import (
	"testing"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/domain"
	"github.com/stretchr/testify/require"
)

func TestIfNoTracksAndAlbumsThenReturnForbidden(t *testing.T) {
	l := mockLibrary{}

	h := music.NewBrowseHandler(l)

	a, err := domain.NewAlbumId("aa")
	require.NoError(t, err)
	b, err := domain.NewAlbumId("bb")
	require.NoError(t, err)

	cmd := music.Browse{
		Ids:        []domain.AlbumId{a, b},
		PublicOnly: false,
	}

	_, err = h.Execute(cmd)
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

func (mockLibrary) Browse(ids []domain.AlbumId, publicOnly bool) (domain.Album, error) {
	return domain.Album{}, nil
}

func (mockLibrary) Search(query string, publicOnly bool) (music.SearchResult, error) {
	return music.SearchResult{}, nil
}

func (mockLibrary) Apply(scan scanner.Album) error {
	return nil
}
