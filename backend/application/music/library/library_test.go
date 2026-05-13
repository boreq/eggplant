package library_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/music/library"
	"github.com/boreq/eggplant/domain"
	"github.com/stretchr/testify/require"
)

// IDs in production are hex (sha256 output). For tests we hex-encode the
// human-readable name so the mock idGenerator's output matches the expected
// values used in test cases.

func nameToHex(name string) string {
	return hex.EncodeToString([]byte(name))
}

func mustAlbumId(name string) domain.AlbumId {
	id, err := domain.NewAlbumId(nameToHex(name))
	if err != nil {
		panic(err)
	}
	return id
}

func mustAlbumTitle(s string) domain.AlbumTitle {
	t, err := domain.NewAlbumTitle(s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustTrackId(name string) domain.TrackId {
	id, err := domain.NewTrackId(nameToHex(name))
	if err != nil {
		panic(err)
	}
	return id
}

func mustTrackTitle(s string) domain.TrackTitle {
	t, err := domain.NewTrackTitle(s)
	if err != nil {
		panic(err)
	}
	return t
}

func mustFileId(name string) domain.FileId {
	id, err := domain.NewFileId(nameToHex(name))
	if err != nil {
		panic(err)
	}
	return id
}

func mustAlbum(id domain.AlbumId, title domain.AlbumTitle, thumbnail *domain.Thumbnail, access domain.Access, parents []domain.AlbumParent, albums []domain.Album, tracks []domain.Track) domain.Album {
	a, err := domain.NewAlbum(id, title, thumbnail, access, parents, albums, tracks)
	if err != nil {
		panic(err)
	}
	return a
}

func rootAlbum(access bool, albums []domain.Album, tracks []domain.Track) domain.Album {
	return mustAlbum(
		domain.AlbumId{},
		mustAlbumTitle("Eggplant"),
		nil,
		domain.NewAccess(access),
		[]domain.AlbumParent{},
		albums,
		tracks,
	)
}

func childAlbum(name string, access bool) domain.Album {
	return mustAlbum(
		mustAlbumId(name),
		mustAlbumTitle(name),
		nil,
		domain.NewAccess(access),
		nil,
		nil,
		nil,
	)
}

func parent(name string) domain.AlbumParent {
	return domain.NewAlbumParent(mustAlbumId(name), mustAlbumTitle(name))
}

func track(name string, fileName string) domain.Track {
	return domain.NewTrack(
		mustTrackId(name),
		mustFileId(fileName),
		mustTrackTitle(name),
		nil,
	)
}

type mockTrackStore struct{}

func (mockTrackStore) SetItems(items []store.Item) {
}

func (mockTrackStore) GetDuration(id string) time.Duration {
	return 0
}

type mockThumbnailStore struct{}

func (mockThumbnailStore) SetItems(items []store.Item) {
}

type mockAccessLoader struct {
	m map[string]domain.Access
}

func (l mockAccessLoader) Load(file string) (domain.Access, error) {
	access, ok := l.m[file]
	if !ok {
		return domain.Access{}, fmt.Errorf("access mapping for '%s' missing", file)
	}
	return access, nil
}

type mockIdGenerator struct{}

func (mockIdGenerator) AlbumId(parents []domain.AlbumId, title string) (domain.AlbumId, error) {
	return domain.NewAlbumId(nameToHex(title))
}

func (mockIdGenerator) TrackId(parents []domain.AlbumId, title string) (domain.TrackId, error) {
	return domain.NewTrackId(nameToHex(title))
}

func (mockIdGenerator) FileId(path string) (domain.FileId, error) {
	return domain.NewFileId(nameToHex(path))
}

func TestLibrary(t *testing.T) {
	expectedNoUpdates := rootAlbum(false, nil, nil)

	expectedListRoot := rootAlbum(false,
		[]domain.Album{childAlbum("a1", true), childAlbum("a2", false)},
		[]domain.Track{track("t1", "t1_path")},
	)

	expectedListChild := mustAlbum(
		mustAlbumId("a1"),
		mustAlbumTitle("a1"),
		nil,
		domain.NewAccess(true),
		[]domain.AlbumParent{parent("a1")},
		[]domain.Album{
			mustAlbum(mustAlbumId("a1a1"), mustAlbumTitle("a1a1"), nil, domain.Access{}, nil, nil, nil),
			childAlbum("a1a2", true),
		},
		[]domain.Track{track("a1t1", "a1t1_path")},
	)

	expectedListRootPublicOnlyDefault := rootAlbum(false,
		[]domain.Album{childAlbum("a1", true)},
		nil,
	)

	expectedListRootPublicOnlyPublic := rootAlbum(true,
		[]domain.Album{childAlbum("a1", true)},
		[]domain.Track{track("t1", "t1_path")},
	)

	expectedListRootOnlyPublic := rootAlbum(false,
		[]domain.Album{childAlbum("a1", true)},
		nil,
	)

	expectedListChildPublicOnly := mustAlbum(
		mustAlbumId("a1"),
		mustAlbumTitle("a1"),
		nil,
		domain.NewAccess(true),
		[]domain.AlbumParent{parent("a1")},
		[]domain.Album{childAlbum("a1a2", true)},
		[]domain.Track{track("a1t1", "a1t1_path")},
	)

	testCases := []struct {
		Name string

		Album      *scanner.Album
		Ids        []domain.AlbumId
		PublicOnly bool
		Access     map[string]domain.Access

		ExpectedAlbum *domain.Album
		ExpectedError error
	}{
		{
			Name:          "no_updates_received",
			Album:         nil,
			Ids:           nil,
			PublicOnly:    false,
			ExpectedAlbum: &expectedNoUpdates,
		},
		{
			Name: "list_root",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				Albums: map[string]*scanner.Album{
					"a1": {AccessFile: "public"},
					"a2": {AccessFile: "no-public"},
				},
				Tracks: map[string]scanner.Track{
					"t1": {Path: "t1_path"},
				},
			},
			Ids:           nil,
			PublicOnly:    false,
			ExpectedAlbum: &expectedListRoot,
		},
		{
			Name: "list_child",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				Albums: map[string]*scanner.Album{
					"a1": {
						AccessFile: "public",
						Albums: map[string]*scanner.Album{
							"a1a1": {AccessFile: "no-public"},
							"a1a2": {AccessFile: "public"},
						},
						Tracks: map[string]scanner.Track{
							"a1t1": {Path: "a1t1_path"},
						},
					},
					"a2": {AccessFile: "no-public"},
				},
				Tracks: map[string]scanner.Track{
					"t1": {Path: "t1_path"},
				},
			},
			Ids:           []domain.AlbumId{mustAlbumId("a1")},
			PublicOnly:    false,
			ExpectedAlbum: &expectedListChild,
		},
		{
			Name: "list_root_public_only_default_access",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				Albums: map[string]*scanner.Album{
					"a1": {AccessFile: "public"},
					"a2": {AccessFile: "no-public"},
				},
				Tracks: map[string]scanner.Track{
					"t1": {Path: "t1_path"},
				},
			},
			Ids:           nil,
			PublicOnly:    true,
			ExpectedAlbum: &expectedListRootPublicOnlyDefault,
		},
		{
			Name: "list_root_public_only_public",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				AccessFile: "public",
				Albums: map[string]*scanner.Album{
					"a1": {AccessFile: "public"},
					"a2": {AccessFile: "no-public"},
				},
				Tracks: map[string]scanner.Track{
					"t1": {Path: "t1_path"},
				},
			},
			Ids:           nil,
			PublicOnly:    true,
			ExpectedAlbum: &expectedListRootPublicOnlyPublic,
		},
		{
			Name: "list_root_only_public",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				Albums: map[string]*scanner.Album{
					"a1": {AccessFile: "public"},
					"a2": {},
				},
				Tracks: map[string]scanner.Track{
					"t1": {Path: "t1_path"},
				},
			},
			Ids:           nil,
			PublicOnly:    true,
			ExpectedAlbum: &expectedListRootOnlyPublic,
		},
		{
			Name: "list_child_public_only",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				Albums: map[string]*scanner.Album{
					"a1": {
						AccessFile: "public",
						Albums: map[string]*scanner.Album{
							"a1a1": {AccessFile: "no-public"},
							"a1a2": {AccessFile: "public"},
						},
						Tracks: map[string]scanner.Track{
							"a1t1": {Path: "a1t1_path"},
						},
					},
					"a2": {AccessFile: "no-public"},
				},
				Tracks: map[string]scanner.Track{
					"t1": {Path: "t1_path"},
				},
			},
			Ids:           []domain.AlbumId{mustAlbumId("a1")},
			PublicOnly:    true,
			ExpectedAlbum: &expectedListChildPublicOnly,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			trs := mockTrackStore{}
			ths := mockThumbnailStore{}
			al := mockAccessLoader{
				m: testCase.Access,
			}
			ig := mockIdGenerator{}

			library, err := library.New(trs, ths, al, ig)
			require.NoError(t, err)

			if testCase.Album != nil {
				require.NoError(t, library.Apply(*testCase.Album))
			}

			album, err := library.Browse(testCase.Ids, testCase.PublicOnly)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedAlbum, &album)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestSearch(t *testing.T) {
	expectedSearchResult := music.SearchResult{
		Tracks: []music.SearchResultTrack{
			{
				Track: domain.NewTrack(
					mustTrackId("album1track1"),
					mustFileId("track1_path"),
					mustTrackTitle("album1track1"),
					nil,
				),
				Album: music.BasicAlbum{
					Title: mustAlbumTitle("album1"),
					Path:  []domain.AlbumId{mustAlbumId("album1")},
				},
			},
		},
		Albums: []music.BasicAlbum{
			{
				Title: mustAlbumTitle("album1"),
				Path:  []domain.AlbumId{mustAlbumId("album1")},
			},
		},
	}

	testCases := []struct {
		Name string

		Album      *scanner.Album
		Query      string
		PublicOnly bool
		Access     map[string]domain.Access

		ExpectedSearchResult music.SearchResult
		ExpectedError        error
	}{
		{
			Name: "find_public_only",
			Access: map[string]domain.Access{
				"public":    domain.NewAccess(true),
				"no-public": domain.NewAccess(false),
			},
			Album: &scanner.Album{
				Albums: map[string]*scanner.Album{
					"album1": {
						AccessFile: "public",
						Tracks: map[string]scanner.Track{
							"album1track1": {Path: "track1_path"},
						},
					},
					"album2": {
						AccessFile: "no-public",
						Tracks: map[string]scanner.Track{
							"album2track1": {Path: "track1_path"},
						},
					},
				},
				Tracks: map[string]scanner.Track{
					"track1": {Path: "track1_path"},
				},
			},
			Query:                "a",
			PublicOnly:           true,
			ExpectedSearchResult: expectedSearchResult,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			trs := mockTrackStore{}
			ths := mockThumbnailStore{}
			al := mockAccessLoader{
				m: testCase.Access,
			}
			ig := mockIdGenerator{}

			library, err := library.New(trs, ths, al, ig)
			require.NoError(t, err)

			if testCase.Album != nil {
				require.NoError(t, library.Apply(*testCase.Album))
			}

			result, err := library.Search(testCase.Query, testCase.PublicOnly)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedSearchResult, result)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestSortTracks(t *testing.T) {
	testCases := []struct {
		Name   string
		Input  []string
		Output []string
	}{
		{
			Name:   "one_word",
			Input:  []string{"ccc", "bbb", "aaa"},
			Output: []string{"aaa", "bbb", "ccc"},
		},
		{
			Name:   "multiple_words",
			Input:  []string{"ccc ccc", "bbb bbb", "aaa aaa"},
			Output: []string{"aaa aaa", "bbb bbb", "ccc ccc"},
		},
		{
			Name:   "numbers",
			Input:  []string{"3", "2", "1"},
			Output: []string{"1", "2", "3"},
		},
		{
			Name:   "same_numbers",
			Input:  []string{"1", "1", "1"},
			Output: []string{"1", "1", "1"},
		},
		{
			Name:   "same_and_words",
			Input:  []string{"1 c", "1 b", "1 a"},
			Output: []string{"1 a", "1 b", "1 c"},
		},
		{
			Name: "numbered",
			Input: []string{
				"10 some title", "9 some title", "8 some title", "7 some title",
				"6 some title", "5 some title", "4 some title", "3 some title",
				"2 some title", "1 some title",
			},
			Output: []string{
				"1 some title", "2 some title", "3 some title", "4 some title",
				"5 some title", "6 some title", "7 some title", "8 some title",
				"9 some title", "10 some title",
			},
		},
		{
			Name: "numbered_dots",
			Input: []string{
				"10. some title", "9. some title", "8. some title", "7. some title",
				"6. some title", "5. some title", "4. some title", "3. some title",
				"2. some title", "1. some title",
			},
			Output: []string{
				"1. some title", "2. some title", "3. some title", "4. some title",
				"5. some title", "6. some title", "7. some title", "8. some title",
				"9. some title", "10. some title",
			},
		},
		{
			Name: "prefixed_numbered",
			Input: []string{
				"10 some title", "09 some title", "08 some title", "07 some title",
				"06 some title", "05 some title", "04 some title", "03 some title",
				"02 some title", "01 some title",
			},
			Output: []string{
				"01 some title", "02 some title", "03 some title", "04 some title",
				"05 some title", "06 some title", "07 some title", "08 some title",
				"09 some title", "10 some title",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var input []domain.Track
			for _, s := range testCase.Input {
				input = append(input, domain.NewTrack(
					domain.TrackId{},
					domain.FileId{},
					mustTrackTitle(s),
					nil,
				))
			}

			var output []domain.Track
			for _, s := range testCase.Output {
				output = append(output, domain.NewTrack(
					domain.TrackId{},
					domain.FileId{},
					mustTrackTitle(s),
					nil,
				))
			}

			library.SortTracks(input)
			require.Equal(t, output, input)
		})
	}
}
