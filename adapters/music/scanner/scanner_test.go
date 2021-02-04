package scanner_test

import (
	"path"
	"testing"

	"github.com/boreq/eggplant/adapters/music/scanner"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	testCases := []struct {
		Name   string
		Result scanner.Album
	}{
		{
			Name: "flat",
			Result: scanner.Album{
				Thumbnail:  "test_data/flat/thumbnail.jpg",
				AccessFile: "test_data/flat/eggplant.access",
				Albums:     map[string]*scanner.Album{},
				Tracks: map[string]scanner.Track{
					"a": {
						Path: "test_data/flat/a.mp3",
					},
					"b": {
						Path: "test_data/flat/b.mp3",
					},
				},
			},
		},
		{
			Name: "one_level",
			Result: scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a": {
						Thumbnail:  "test_data/one_level/a/thumbnail.jpg",
						AccessFile: "",
						Albums:     map[string]*scanner.Album{},
						Tracks: map[string]scanner.Track{
							"a": {
								Path: "test_data/one_level/a/a.mp3",
							},
							"b": {
								Path: "test_data/one_level/a/b.mp3",
							},
						},
					},
					"b": {
						Thumbnail:  "",
						AccessFile: "test_data/one_level/b/eggplant.access",
						Albums:     map[string]*scanner.Album{},
						Tracks: map[string]scanner.Track{
							"a": {
								Path: "test_data/one_level/b/a.mp3",
							},
							"b": {
								Path: "test_data/one_level/b/b.mp3",
							},
						},
					},
					"c": {
						Thumbnail:  "",
						AccessFile: "",
						Albums:     map[string]*scanner.Album{},
						Tracks: map[string]scanner.Track{
							"a": {
								Path: "test_data/one_level/c/a.mp3",
							},
							"b": {
								Path: "test_data/one_level/c/b.mp3",
							},
						},
					},
				},
				Tracks: map[string]scanner.Track{},
			},
		},
		{
			Name: "multiple_levels",
			Result: scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a": {
						Thumbnail:  "test_data/multiple_levels/a/thumbnail.jpg",
						AccessFile: "",
						Albums: map[string]*scanner.Album{
							"a": {
								Thumbnail:  "test_data/multiple_levels/a/a/thumbnail.jpg",
								AccessFile: "",
								Albums:     map[string]*scanner.Album{},
								Tracks: map[string]scanner.Track{
									"a": {
										Path: "test_data/multiple_levels/a/a/a.mp3",
									},
									"b": {
										Path: "test_data/multiple_levels/a/a/b.mp3",
									},
								},
							},
							"b": {
								Thumbnail:  "",
								AccessFile: "test_data/multiple_levels/a/b/eggplant.access",
								Albums:     map[string]*scanner.Album{},
								Tracks: map[string]scanner.Track{
									"a": {
										Path: "test_data/multiple_levels/a/b/a.mp3",
									},
									"b": {
										Path: "test_data/multiple_levels/a/b/b.mp3",
									},
								},
							},
						},
						Tracks: map[string]scanner.Track{
							"a": {
								Path: "test_data/multiple_levels/a/a.mp3",
							},
							"b": {
								Path: "test_data/multiple_levels/a/b.mp3",
							},
						},
					},
					"b": {
						Thumbnail:  "",
						AccessFile: "",
						Albums: map[string]*scanner.Album{
							"a": {
								Thumbnail:  "test_data/multiple_levels/b/a/thumbnail.jpg",
								AccessFile: "",
								Albums:     map[string]*scanner.Album{},
								Tracks: map[string]scanner.Track{
									"a": {
										Path: "test_data/multiple_levels/b/a/a.mp3",
									},
									"b": {
										Path: "test_data/multiple_levels/b/a/b.mp3",
									},
								},
							},
							"b": {
								Thumbnail:  "",
								AccessFile: "test_data/multiple_levels/b/b/eggplant.access",
								Albums:     map[string]*scanner.Album{},
								Tracks: map[string]scanner.Track{
									"a": {
										Path: "test_data/multiple_levels/b/b/a.mp3",
									},
									"b": {
										Path: "test_data/multiple_levels/b/b/b.mp3",
									},
								},
							},
						},
						Tracks: map[string]scanner.Track{},
					},
				},
				Tracks: map[string]scanner.Track{},
			},
		},
		{
			Name: "symlinks",
			Result: scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums:     map[string]*scanner.Album{},
				Tracks:     map[string]scanner.Track{},
			},
		},
		{
			Name: "mixed",
			Result: scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"mixed": {
						Thumbnail:  "",
						AccessFile: "",
						Albums: map[string]*scanner.Album{
							"songs": {
								Thumbnail:  "test_data/mixed/mixed/songs/thumbnail.jpg",
								AccessFile: "",
								Albums:     map[string]*scanner.Album{},
								Tracks: map[string]scanner.Track{
									"a": {
										Path: "test_data/mixed/mixed/songs/a.mp3",
									},
									"b": {
										Path: "test_data/mixed/mixed/songs/b.mp3",
									},
								},
							},
						},
						Tracks: map[string]scanner.Track{},
					},
					"songs": {
						Thumbnail:  "test_data/mixed/songs/thumbnail.jpg",
						AccessFile: "",
						Albums:     map[string]*scanner.Album{},
						Tracks: map[string]scanner.Track{

							"a": {
								Path: "test_data/mixed/songs/a.mp3",
							},
							"b": {
								Path: "test_data/mixed/songs/b.mp3",
							},
						},
					},
				},
				Tracks: map[string]scanner.Track{},
			},
		},
		{
			Name: "some_empty",
			Result: scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums: map[string]*scanner.Album{
					"a": {
						Thumbnail:  "",
						AccessFile: "",
						Albums: map[string]*scanner.Album{
							"a": {
								Thumbnail:  "test_data/some_empty/a/a/thumbnail.jpg",
								AccessFile: "",
								Albums:     map[string]*scanner.Album{},
								Tracks: map[string]scanner.Track{
									"a": {
										Path: "test_data/some_empty/a/a/a.mp3",
									},
									"b": {
										Path: "test_data/some_empty/a/a/b.mp3",
									},
								},
							},
						},
						Tracks: map[string]scanner.Track{},
					},
				},
				Tracks: map[string]scanner.Track{},
			},
		},
		{
			Name: "case_insensitive_extensions",
			Result: scanner.Album{
				Thumbnail:  "",
				AccessFile: "",
				Albums:     map[string]*scanner.Album{},
				Tracks: map[string]scanner.Track{
					"a": {
						Path: "test_data/case_insensitive_extensions/a.mp3",
					},
					"b": {
						Path: "test_data/case_insensitive_extensions/b.MP3",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			config := scanner.Config{
				TrackExtensions: []string{
					".mp3",
				},
			}

			s, err := scanner.New(testDirectory(testCase.Name), config)
			require.NoError(t, err)

			c, err := s.Start()
			require.NoError(t, err)

			album := <-c
			require.Equal(t, testCase.Result, album)

		})
	}
}

func testDirectory(name string) string {
	return path.Join("test_data", name)
}
