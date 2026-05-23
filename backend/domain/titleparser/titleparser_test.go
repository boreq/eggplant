package titleparser_test

import (
	"testing"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/eggplant/domain/titleparser"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		number int
		has    bool
		title  string
	}{
		{
			name:   "dash_separator",
			input:  "01 - Music Song",
			number: 1,
			has:    true,
			title:  "Music Song",
		},
		{
			name:   "space_separator",
			input:  "01 Music Song",
			number: 1,
			has:    true,
			title:  "Music Song",
		},
		{
			name:   "dot_separator",
			input:  "01. Music Song",
			number: 1,
			has:    true,
			title:  "Music Song",
		},
		{
			name:  "no_number",
			input: "Just a title",
			has:   false,
			title: "Just a title",
		},
		{
			name:   "multi_digit_number",
			input:  "100 Track",
			number: 100,
			has:    true,
			title:  "Track",
		},
		{
			name:  "no_separator",
			input: "01Title",
			has:   false,
			title: "01Title",
		},
		{
			name:   "leading_whitespace",
			input:  "  01 - Song",
			number: 1,
			has:    true,
			title:  "Song",
		},
		{
			name:   "leading_zero",
			input:  "007 - Bond",
			number: 7,
			has:    true,
			title:  "Bond",
		},
		{
			name:   "zero",
			input:  "0 - Intro",
			number: 0,
			has:    true,
			title:  "Intro",
		},
		{
			name:   "zero_padded",
			input:  "00. Hidden Track",
			number: 0,
			has:    true,
			title:  "Hidden Track",
		},
		{
			name:   "mixed_separators",
			input:  "01.- Hyphenated",
			number: 1,
			has:    true,
			title:  "Hyphenated",
		},
		{
			name:   "title_starts_with_letter_after_number",
			input:  "01-Title",
			number: 1,
			has:    true,
			title:  "Title",
		},
		{
			name:   "title_with_numbers_inside",
			input:  "5. Track 1 of many",
			number: 5,
			has:    true,
			title:  "Track 1 of many",
		},
		{
			name:  "title_is_just_a_number",
			input: "42",
			has:   false,
			title: "42",
		},
		{
			name:  "title_is_just_a_zero",
			input: "0",
			has:   false,
			title: "0",
		},
		{
			name:  "only_whitespace_falls_back",
			input: "   ",
			has:   false,
			title: "   ",
		},
		{
			name:  "number_overflow_falls_back",
			input: "99999999999999999999 - Title",
			has:   false,
			title: "99999999999999999999 - Title",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := titleparser.Parse(tc.input)
			require.NoError(t, err)

			number := got.Number()
			if tc.has {
				require.NotNil(t, number, "number")
				require.Equal(t, tc.number, number.Int(), "number")
			} else {
				require.Nil(t, number, "number")
			}

			expectedTitle, err := domain.NewTrackTitle(tc.title)
			require.NoError(t, err)
			require.Equal(t, expectedTitle, got.Title(), "title")
		})
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := titleparser.Parse("")
	require.Error(t, err)
}
