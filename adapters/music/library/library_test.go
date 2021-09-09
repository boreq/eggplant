package library_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/music/library"
	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/adapters/music/store"
	"github.com/boreq/eggplant/application/music"
	"github.com/stretchr/testify/require"
)

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
	m map[string]music.Access
}

func (l mockAccessLoader) Load(file string) (music.Access, error) {
	access, ok := l.m[file]
	if !ok {
		return music.Access{}, fmt.Errorf("access mapping for '%s' missing", file)
	}
	return access, nil
}

type mockIdGenerator struct{}

func (mockIdGenerator) AlbumId(parents []music.AlbumId, title string) (music.AlbumId, error) {
	return music.AlbumId(title), nil
}

func (mockIdGenerator) TrackId(parents []music.AlbumId, title string) (music.TrackId, error) {
	return music.TrackId(title), nil
}

func (mockIdGenerator) FileId(path string) (music.FileId, error) {
	return music.FileId(path), nil
}

func TestLibrary(t *testing.T) {
	testCases := []struct {
		Name string

		Album      *scanner.Album
		Ids        []music.AlbumId
		PublicOnly bool
		Access     map[string]music.Access

		ExpectedAlbum *music.Album
		ExpectedError error
	}{
		{
			Name:       "no_updates_received",
			Album:      nil,
			Ids:        nil,
			PublicOnly: false,
			ExpectedAlbum: &music.Album{
				Id:        "",
				Title:     "Eggplant",
				Thumbnail: nil,
				Access: music.Access{
					Public: false,
				},
				Parents: []music.Album{},
				Albums:  nil,
				Tracks:  nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "list_root",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a1": &scanner.Album{
						AccessFile: "public",
					},
					"a2": &scanner.Album{
						AccessFile: "no-public",
					},
				},
				Tracks: map[string]scanner.Track{
					"t1": scanner.Track{
						Path: "t1_path",
					},
				},
			},
			Ids:        nil,
			PublicOnly: false,
			ExpectedAlbum: &music.Album{
				Id:        "",
				Title:     "Eggplant",
				Thumbnail: nil,
				Access: music.Access{
					Public: false,
				},
				Parents: []music.Album{},
				Albums: []music.Album{
					{
						Title: "a1",
						Id:    "a1",
						Access: music.Access{
							Public: true,
						},
					},
					{
						Title: "a2",
						Id:    "a2",
						Access: music.Access{
							Public: false,
						},
					},
				},
				Tracks: []music.Track{
					{
						Id:     "t1",
						Title:  "t1",
						FileId: "t1_path",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "list_child",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a1": &scanner.Album{
						AccessFile: "public",
						Albums: map[string]*scanner.Album{
							"a1a1": &scanner.Album{
								AccessFile: "no-public",
							},
							"a1a2": &scanner.Album{
								AccessFile: "public",
							},
						},
						Tracks: map[string]scanner.Track{
							"a1t1": scanner.Track{
								Path: "a1t1_path",
							},
						},
					},
					"a2": &scanner.Album{
						AccessFile: "no-public",
					},
				},
				Tracks: map[string]scanner.Track{
					"t1": scanner.Track{
						Path: "t1_path",
					},
				},
			},
			Ids:        []music.AlbumId{"a1"},
			PublicOnly: false,
			ExpectedAlbum: &music.Album{
				Id:        "a1",
				Title:     "a1",
				Thumbnail: nil,
				Access: music.Access{
					Public: true,
				},
				Parents: []music.Album{
					{
						Id:    "a1",
						Title: "a1",
					},
				},
				Albums: []music.Album{
					{
						Title: "a1a1",
						Id:    "a1a1",
					},
					{
						Title: "a1a2",
						Id:    "a1a2",
						Access: music.Access{
							Public: true,
						},
					},
				},
				Tracks: []music.Track{
					{
						Id:     "a1t1",
						Title:  "a1t1",
						FileId: "a1t1_path",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "list_root_public_only_default_access",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a1": &scanner.Album{
						AccessFile: "public",
					},
					"a2": &scanner.Album{
						AccessFile: "no-public",
					},
				},
				Tracks: map[string]scanner.Track{
					"t1": scanner.Track{
						Path: "t1_path",
					},
				},
			},
			Ids:        nil,
			PublicOnly: true,
			ExpectedAlbum: &music.Album{
				Id:        "",
				Title:     "Eggplant",
				Thumbnail: nil,
				Access: music.Access{
					Public: false,
				},
				Parents: []music.Album{},
				Albums: []music.Album{
					{
						Title: "a1",
						Id:    "a1",
						Access: music.Access{
							Public: true,
						},
					},
				},
				Tracks: nil,
			},
			ExpectedError: nil,
		},
		{
			Name: "list_root_public_only_public",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail:  "",
				AccessFile: "public",
				Albums: map[string]*scanner.Album{
					"a1": &scanner.Album{
						AccessFile: "public",
					},
					"a2": &scanner.Album{
						AccessFile: "no-public",
					},
				},
				Tracks: map[string]scanner.Track{
					"t1": scanner.Track{
						Path: "t1_path",
					},
				},
			},
			Ids:        nil,
			PublicOnly: true,
			ExpectedAlbum: &music.Album{
				Id:        "",
				Title:     "Eggplant",
				Thumbnail: nil,
				Access: music.Access{
					Public: true,
				},
				Parents: []music.Album{},
				Albums: []music.Album{
					{
						Title: "a1",
						Id:    "a1",
						Access: music.Access{
							Public: true,
						},
					},
				},
				Tracks: []music.Track{
					{
						Id:     "t1",
						Title:  "t1",
						FileId: "t1_path",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "list_root_only_public",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail: "",
				Albums: map[string]*scanner.Album{
					"a1": &scanner.Album{
						AccessFile: "public",
					},
					"a2": &scanner.Album{},
				},
				Tracks: map[string]scanner.Track{
					"t1": scanner.Track{
						Path: "t1_path",
					},
				},
			},
			Ids:        nil,
			PublicOnly: true,
			ExpectedAlbum: &music.Album{
				Id:        "",
				Title:     "Eggplant",
				Thumbnail: nil,
				Access: music.Access{
					Public: false,
				},
				Parents: []music.Album{},
				Albums: []music.Album{
					{
						Title: "a1",
						Id:    "a1",
						Access: music.Access{
							Public: true,
						},
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "list_child_public_only",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a1": &scanner.Album{
						AccessFile: "public",
						Albums: map[string]*scanner.Album{
							"a1a1": &scanner.Album{
								AccessFile: "no-public",
							},
							"a1a2": &scanner.Album{
								AccessFile: "public",
							},
						},
						Tracks: map[string]scanner.Track{
							"a1t1": scanner.Track{
								Path: "a1t1_path",
							},
						},
					},
					"a2": &scanner.Album{
						AccessFile: "no-public",
					},
				},
				Tracks: map[string]scanner.Track{
					"t1": scanner.Track{
						Path: "t1_path",
					},
				},
			},
			Ids:        []music.AlbumId{"a1"},
			PublicOnly: true,
			ExpectedAlbum: &music.Album{
				Id:        "a1",
				Title:     "a1",
				Thumbnail: nil,
				Access: music.Access{
					Public: true,
				},
				Parents: []music.Album{
					{
						Id:    "a1",
						Title: "a1",
					},
				},
				Albums: []music.Album{
					{
						Title: "a1a2",
						Id:    "a1a2",
						Access: music.Access{
							Public: true,
						},
					},
				},
				Tracks: []music.Track{
					{
						Id:     "a1t1",
						Title:  "a1t1",
						FileId: "a1t1_path",
					},
				},
			},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ch := make(chan scanner.Album)
			trs := mockTrackStore{}
			ths := mockThumbnailStore{}
			al := mockAccessLoader{
				m: testCase.Access,
			}
			ig := mockIdGenerator{}

			library, err := library.New(ch, trs, ths, al, ig)
			require.NoError(t, err)

			if testCase.Album != nil {
				ch <- *testCase.Album
				<-time.After(time.Second) // rc
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
	testCases := []struct {
		Name string

		Album      *scanner.Album
		Query      string
		PublicOnly bool
		Access     map[string]music.Access

		ExpectedSearchResult music.SearchResult
		ExpectedError        error
	}{
		{
			Name: "find_public_only",
			Access: map[string]music.Access{
				"public": {
					Public: true,
				},
				"no-public": {
					Public: false,
				},
			},
			Album: &scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"album1": &scanner.Album{
						AccessFile: "public",
						Tracks: map[string]scanner.Track{
							"album1track1": scanner.Track{
								Path: "track1_path",
							},
						},
					},
					"album2": &scanner.Album{
						AccessFile: "no-public",
						Tracks: map[string]scanner.Track{
							"album2track1": scanner.Track{
								Path: "track1_path",
							},
						},
					},
				},
				Tracks: map[string]scanner.Track{
					"track1": scanner.Track{
						Path: "track1_path",
					},
				},
			},
			Query:      "a",
			PublicOnly: true,
			ExpectedSearchResult: music.SearchResult{
				Tracks: []music.SearchResultTrack{
					{
						Track: music.Track{
							Id:     "album1track1",
							FileId: "track1_path",
							Title:  "album1track1",
						},
						Album: music.BasicAlbum{
							Title: "album1",
							Path: []music.AlbumId{
								"album1",
							},
						},
					},
				},
				Albums: []music.BasicAlbum{
					{
						Title: "album1",
						Path: []music.AlbumId{
							"album1",
						},
					},
				},
			},
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ch := make(chan scanner.Album)
			trs := mockTrackStore{}
			ths := mockThumbnailStore{}
			al := mockAccessLoader{
				m: testCase.Access,
			}
			ig := mockIdGenerator{}

			library, err := library.New(ch, trs, ths, al, ig)
			require.NoError(t, err)

			if testCase.Album != nil {
				ch <- *testCase.Album
				<-time.After(time.Second) // rc
			}

			result, err := library.Search(
				testCase.Query,
				testCase.PublicOnly,
			)
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
			Name: "one_word",
			Input: []string{
				"ccc",
				"bbb",
				"aaa",
			},
			Output: []string{
				"aaa",
				"bbb",
				"ccc",
			},
		},
		{
			Name: "multiple_words",
			Input: []string{
				"ccc ccc",
				"bbb bbb",
				"aaa aaa",
			},
			Output: []string{
				"aaa aaa",
				"bbb bbb",
				"ccc ccc",
			},
		},
		{
			Name: "numbers",
			Input: []string{
				"3",
				"2",
				"1",
			},
			Output: []string{
				"1",
				"2",
				"3",
			},
		},
		{
			Name: "same_numbers",
			Input: []string{
				"1",
				"1",
				"1",
			},
			Output: []string{
				"1",
				"1",
				"1",
			},
		},
		{
			Name: "same_and_words",
			Input: []string{
				"1 c",
				"1 b",
				"1 a",
			},
			Output: []string{
				"1 a",
				"1 b",
				"1 c",
			},
		},
		{
			Name: "numbered",
			Input: []string{
				"10 some title",
				"9 some title",
				"8 some title",
				"7 some title",
				"6 some title",
				"5 some title",
				"4 some title",
				"3 some title",
				"2 some title",
				"1 some title",
			},
			Output: []string{
				"1 some title",
				"2 some title",
				"3 some title",
				"4 some title",
				"5 some title",
				"6 some title",
				"7 some title",
				"8 some title",
				"9 some title",
				"10 some title",
			},
		},
		{
			Name: "numbered_dots",
			Input: []string{
				"10. some title",
				"9. some title",
				"8. some title",
				"7. some title",
				"6. some title",
				"5. some title",
				"4. some title",
				"3. some title",
				"2. some title",
				"1. some title",
			},
			Output: []string{
				"1. some title",
				"2. some title",
				"3. some title",
				"4. some title",
				"5. some title",
				"6. some title",
				"7. some title",
				"8. some title",
				"9. some title",
				"10. some title",
			},
		},
		{
			Name: "prefixed_numbered",
			Input: []string{
				"10 some title",
				"09 some title",
				"08 some title",
				"07 some title",
				"06 some title",
				"05 some title",
				"04 some title",
				"03 some title",
				"02 some title",
				"01 some title",
			},
			Output: []string{
				"01 some title",
				"02 some title",
				"03 some title",
				"04 some title",
				"05 some title",
				"06 some title",
				"07 some title",
				"08 some title",
				"09 some title",
				"10 some title",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var input []music.Track
			for _, s := range testCase.Input {
				input = append(input, music.Track{Title: s})
			}

			var output []music.Track
			for _, s := range testCase.Output {
				output = append(output, music.Track{Title: s})
			}

			library.SortTracks(input)
			require.Equal(t, output, input)
		})
	}
}
