package music_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/titleparser"
	"github.com/stretchr/testify/require"
)

func TestNewTracks(t *testing.T) {
	cases := []struct {
		name   string
		input  []string
		output []string
	}{
		{
			name: "one_word",
			input: []string{
				"ccc",
				"bbb",
				"aaa",
			},
			output: []string{
				"aaa",
				"bbb",
				"ccc",
			},
		},
		{
			name: "multiple_words",
			input: []string{
				"ccc ccc",
				"bbb bbb",
				"aaa aaa",
			},
			output: []string{
				"aaa aaa",
				"bbb bbb",
				"ccc ccc",
			},
		},
		{
			name: "numbers",
			input: []string{
				"3",
				"2",
				"1",
			},
			output: []string{
				"1",
				"2",
				"3",
			},
		},
		{
			name: "same_numbers",
			input: []string{
				"1",
				"1",
				"1",
			},
			output: []string{
				"1",
				"1",
				"1",
			},
		},
		{
			name: "same_and_words",
			input: []string{
				"1 c",
				"1 b",
				"1 a",
			},
			output: []string{
				"1 a",
				"1 b",
				"1 c",
			},
		},
		{
			name: "numbered",
			input: []string{
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
			output: []string{
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
			name: "numbered_dots",
			input: []string{
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
			output: []string{
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
			name: "prefixed_numbered",
			input: []string{
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
			output: []string{
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := mkTracks(t, tc.input)
			expected := mkTracks(t, tc.output)
			got := music.NewTracks(input)
			require.Equal(t, expected, got.Items())
		})
	}
}

func mkTracks(t *testing.T, titles []string) []music.Track {
	t.Helper()
	dur, err := music.NewTrackDuration(time.Second)
	require.NoError(t, err)
	tracks := make([]music.Track, 0, len(titles))
	for _, s := range titles {
		title, err := music.NewTrackTitle(s)
		require.NoError(t, err)
		parsed, err := titleparser.Parse(title)
		require.NoError(t, err)
		if parsed.Number() == nil {
			tracks = append(tracks, music.NewTrack(music.TrackId{}, music.FileId{}, parsed.Title(), dur))
		} else {
			tracks = append(tracks, music.NewTrackWithNumber(music.TrackId{}, music.FileId{}, *parsed.Number(), parsed.Title(), dur))
		}
	}
	return tracks
}
