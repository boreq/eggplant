package scanner_test

import (
	"path"
	"testing"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/boreq/eggplant/domain/music"
	scannerdomain "github.com/boreq/eggplant/domain/music/scanner"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
)

type testTrack struct {
	title string
	path  string
}

type testAlbum struct {
	title      string
	thumbnail  string
	accessFile string
	albums     []testAlbum
	tracks     []testTrack
}

func TestScanner(t *testing.T) {
	testCases := []struct {
		Name   string
		Result *testAlbum
		Error  error
	}{
		{
			Name: "flat",
			Result: &testAlbum{
				thumbnail:  "test_data/flat/thumbnail.jpg",
				accessFile: "test_data/flat/eggplant.access",
				tracks: []testTrack{
					{title: "a", path: "test_data/flat/a.mp3"},
					{title: "b", path: "test_data/flat/b.mp3"},
				},
			},
		},
		{
			Name: "one_level",
			Result: &testAlbum{
				albums: []testAlbum{
					{
						title:     "a",
						thumbnail: "test_data/one_level/a/thumbnail.jpg",
						tracks: []testTrack{
							{title: "a", path: "test_data/one_level/a/a.mp3"},
							{title: "b", path: "test_data/one_level/a/b.mp3"},
						},
					},
					{
						title:      "b",
						accessFile: "test_data/one_level/b/eggplant.access",
						tracks: []testTrack{
							{title: "a", path: "test_data/one_level/b/a.mp3"},
							{title: "b", path: "test_data/one_level/b/b.mp3"},
						},
					},
					{
						title: "c",
						tracks: []testTrack{
							{title: "a", path: "test_data/one_level/c/a.mp3"},
							{title: "b", path: "test_data/one_level/c/b.mp3"},
						},
					},
				},
			},
		},
		{
			Name: "multiple_levels",
			Result: &testAlbum{
				albums: []testAlbum{
					{
						title:     "a",
						thumbnail: "test_data/multiple_levels/a/thumbnail.jpg",
						albums: []testAlbum{
							{
								title:     "a",
								thumbnail: "test_data/multiple_levels/a/a/thumbnail.jpg",
								tracks: []testTrack{
									{title: "a", path: "test_data/multiple_levels/a/a/a.mp3"},
									{title: "b", path: "test_data/multiple_levels/a/a/b.mp3"},
								},
							},
							{
								title:      "b",
								accessFile: "test_data/multiple_levels/a/b/eggplant.access",
								tracks: []testTrack{
									{title: "a", path: "test_data/multiple_levels/a/b/a.mp3"},
									{title: "b", path: "test_data/multiple_levels/a/b/b.mp3"},
								},
							},
						},
						tracks: []testTrack{
							{title: "a", path: "test_data/multiple_levels/a/a.mp3"},
							{title: "b", path: "test_data/multiple_levels/a/b.mp3"},
						},
					},
					{
						title: "b",
						albums: []testAlbum{
							{
								title:     "a",
								thumbnail: "test_data/multiple_levels/b/a/thumbnail.jpg",
								tracks: []testTrack{
									{title: "a", path: "test_data/multiple_levels/b/a/a.mp3"},
									{title: "b", path: "test_data/multiple_levels/b/a/b.mp3"},
								},
							},
							{
								title:      "b",
								accessFile: "test_data/multiple_levels/b/b/eggplant.access",
								tracks: []testTrack{
									{title: "a", path: "test_data/multiple_levels/b/b/a.mp3"},
									{title: "b", path: "test_data/multiple_levels/b/b/b.mp3"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "symlinks",
			Result: &testAlbum{
				albums: []testAlbum{
					{
						title:     "a",
						thumbnail: "test_data/symlinks/a/thumbnail.jpg",
						tracks: []testTrack{
							{title: "a", path: "test_data/symlinks/a/a.mp3"},
							{title: "b", path: "test_data/symlinks/a/b.mp3"},
						},
					},
					{
						title:      "b",
						accessFile: "test_data/symlinks/b/eggplant.access",
						tracks: []testTrack{
							{title: "a", path: "test_data/symlinks/b/a.mp3"},
							{title: "b", path: "test_data/symlinks/b/b.mp3"},
						},
					},
					{
						title: "c",
						tracks: []testTrack{
							{title: "a", path: "test_data/symlinks/c/a.mp3"},
							{title: "b", path: "test_data/symlinks/c/b.mp3"},
						},
					},
				},
				tracks: []testTrack{
					{title: "a", path: "test_data/symlinks/a.mp3"},
					{title: "b", path: "test_data/symlinks/b.mp3"},
				},
			},
		},
		{
			Name: "mixed",
			Result: &testAlbum{
				albums: []testAlbum{
					{
						title: "mixed",
						albums: []testAlbum{
							{
								title:     "songs",
								thumbnail: "test_data/mixed/mixed/songs/thumbnail.jpg",
								tracks: []testTrack{
									{title: "a", path: "test_data/mixed/mixed/songs/a.mp3"},
									{title: "b", path: "test_data/mixed/mixed/songs/b.mp3"},
								},
							},
						},
					},
					{
						title:     "songs",
						thumbnail: "test_data/mixed/songs/thumbnail.jpg",
						tracks: []testTrack{
							{title: "a", path: "test_data/mixed/songs/a.mp3"},
							{title: "b", path: "test_data/mixed/songs/b.mp3"},
						},
					},
				},
			},
		},
		{
			Name: "some_empty",
			Result: &testAlbum{
				albums: []testAlbum{
					{
						title: "a",
						albums: []testAlbum{
							{
								title:     "a",
								thumbnail: "test_data/some_empty/a/a/thumbnail.jpg",
								tracks: []testTrack{
									{title: "a", path: "test_data/some_empty/a/a/a.mp3"},
									{title: "b", path: "test_data/some_empty/a/a/b.mp3"},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "case_insensitive_extensions",
			Result: &testAlbum{
				tracks: []testTrack{
					{title: "a", path: "test_data/case_insensitive_extensions/a.mp3"},
					{title: "b", path: "test_data/case_insensitive_extensions/b.MP3"},
				},
			},
		},
		{
			Name: "case_insensitive_thumbnails",
			Result: &testAlbum{
				albums: []testAlbum{
					{
						title:     "a",
						thumbnail: "test_data/case_insensitive_thumbnails/a/thumbnail.jpg",
						tracks: []testTrack{
							{title: "a", path: "test_data/case_insensitive_thumbnails/a/a.mp3"},
							{title: "b", path: "test_data/case_insensitive_thumbnails/a/b.mp3"},
						},
					},
					{
						title:     "b",
						thumbnail: "test_data/case_insensitive_thumbnails/b/THUMBNAIL.JPG",
						tracks: []testTrack{
							{title: "a", path: "test_data/case_insensitive_thumbnails/b/a.mp3"},
							{title: "b", path: "test_data/case_insensitive_thumbnails/b/b.mp3"},
						},
					},
				},
			},
		},
		{
			Name:  "symlinks_loop",
			Error: errors.New("walk failed: loop detected: 'test_data/symlinks_loop/a' visited multiple times"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			s, err := scanner.New(testDirectory(testCase.Name), testConfig(t))
			require.NoError(t, err)

			album, err := s.Scan()

			if testCase.Error == nil {
				require.NoError(t, err)
				require.Equal(t, buildRootAlbum(t, testCase.Result), album)
			} else {
				require.EqualError(t, err, testCase.Error.Error())
			}
		})
	}
}

func TestScannerFailsIfDirectoryDoesNotExist(t *testing.T) {
	s, err := scanner.New("some-completely-made-up-file-name-come-on-surely-this-does-not-exist", testConfig(t))
	require.NoError(t, err)

	_, err = s.Scan()
	require.EqualError(t, err, "walk failed: received an error: lstat some-completely-made-up-file-name-come-on-surely-this-does-not-exist: no such file or directory")
}

func testConfig(t *testing.T) scanner.Config {
	t.Helper()

	trackExt, err := music.NewFileExtension(".mp3")
	require.NoError(t, err)

	thumbnailStem, err := scanner.NewThumbnailStem("thumbnail")
	require.NoError(t, err)

	thumbnailExt, err := music.NewFileExtension(".jpg")
	require.NoError(t, err)

	cfg, err := scanner.NewConfig(
		[]music.FileExtension{trackExt},
		[]scanner.ThumbnailStem{thumbnailStem},
		[]music.FileExtension{thumbnailExt},
	)
	require.NoError(t, err)
	return cfg
}

func testDirectory(name string) string {
	return path.Join("test_data", name)
}

func buildRootAlbum(t *testing.T, src *testAlbum) scannerdomain.FoundRootAlbum {
	t.Helper()
	return scannerdomain.NewFoundRootAlbum(
		optionalFilePath(t, src.thumbnail),
		optionalFilePath(t, src.accessFile),
		buildAlbums(t, src.albums),
		buildTracks(t, src.tracks),
	)
}

func buildAlbums(t *testing.T, src []testAlbum) map[music.AlbumTitle]scannerdomain.FoundAlbum {
	t.Helper()
	out := map[music.AlbumTitle]scannerdomain.FoundAlbum{}
	for _, a := range src {
		title, err := music.NewAlbumTitle(a.title)
		require.NoError(t, err)

		album, err := scannerdomain.NewFoundAlbum(
			optionalFilePath(t, a.thumbnail),
			optionalFilePath(t, a.accessFile),
			buildAlbums(t, a.albums),
			buildTracks(t, a.tracks),
		)
		require.NoError(t, err)
		out[title] = album
	}
	return out
}

func buildTracks(t *testing.T, src []testTrack) map[music.TrackTitle]scannerdomain.FoundTrack {
	t.Helper()
	out := map[music.TrackTitle]scannerdomain.FoundTrack{}
	for _, tr := range src {
		title, err := music.NewTrackTitle(tr.title)
		require.NoError(t, err)

		p, err := music.NewFilePath(tr.path)
		require.NoError(t, err)

		out[title] = scannerdomain.NewFoundTrack(p)
	}
	return out
}

func optionalFilePath(t *testing.T, raw string) *music.FilePath {
	t.Helper()
	if raw == "" {
		return nil
	}
	p, err := music.NewFilePath(raw)
	require.NoError(t, err)
	return &p
}
