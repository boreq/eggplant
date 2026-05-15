package domain_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/domain"
	"github.com/stretchr/testify/require"
)

func TestSortTracks(t *testing.T) {
	cases := []struct {
		name   string
		input  []string
		output []string
	}{
		{
			name:   "one_word",
			input:  []string{"ccc", "bbb", "aaa"},
			output: []string{"aaa", "bbb", "ccc"},
		},
		{
			name:   "multiple_words",
			input:  []string{"ccc ccc", "bbb bbb", "aaa aaa"},
			output: []string{"aaa aaa", "bbb bbb", "ccc ccc"},
		},
		{
			name:   "numbers",
			input:  []string{"3", "2", "1"},
			output: []string{"1", "2", "3"},
		},
		{
			name:   "same_numbers",
			input:  []string{"1", "1", "1"},
			output: []string{"1", "1", "1"},
		},
		{
			name:   "same_and_words",
			input:  []string{"1 c", "1 b", "1 a"},
			output: []string{"1 a", "1 b", "1 c"},
		},
		{
			name: "numbered",
			input: []string{
				"10 some title", "9 some title", "8 some title", "7 some title",
				"6 some title", "5 some title", "4 some title", "3 some title",
				"2 some title", "1 some title",
			},
			output: []string{
				"1 some title", "2 some title", "3 some title", "4 some title",
				"5 some title", "6 some title", "7 some title", "8 some title",
				"9 some title", "10 some title",
			},
		},
		{
			name: "numbered_dots",
			input: []string{
				"10. some title", "9. some title", "8. some title", "7. some title",
				"6. some title", "5. some title", "4. some title", "3. some title",
				"2. some title", "1. some title",
			},
			output: []string{
				"1. some title", "2. some title", "3. some title", "4. some title",
				"5. some title", "6. some title", "7. some title", "8. some title",
				"9. some title", "10. some title",
			},
		},
		{
			name: "prefixed_numbered",
			input: []string{
				"10 some title", "09 some title", "08 some title", "07 some title",
				"06 some title", "05 some title", "04 some title", "03 some title",
				"02 some title", "01 some title",
			},
			output: []string{
				"01 some title", "02 some title", "03 some title", "04 some title",
				"05 some title", "06 some title", "07 some title", "08 some title",
				"09 some title", "10 some title",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := mkTracks(t, tc.input)
			expected := mkTracks(t, tc.output)
			domain.SortTracks(input)
			require.Equal(t, expected, input)
		})
	}
}

func mkTracks(t *testing.T, titles []string) []domain.Track {
	t.Helper()
	dur, err := domain.NewTrackDuration(time.Second)
	require.NoError(t, err)
	tracks := make([]domain.Track, 0, len(titles))
	for _, s := range titles {
		title, err := domain.NewTrackTitle(s)
		require.NoError(t, err)
		tracks = append(tracks, domain.NewTrack(domain.TrackId{}, domain.FileId{}, title, dur))
	}
	return tracks
}
