package library_test

import (
	"os"
	"testing"

	"github.com/boreq/eggplant/adapters/music/library"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
)

func TestAccessLoaderYes(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         string
		ExpectedError string
		ResultPublic  bool
	}{
		{
			Name:         "yes",
			Input:        "public: yes",
			ResultPublic: true,
		},
		{
			Name:         "no",
			Input:        "public: no",
			ResultPublic: false,
		},
		{
			Name:          "duplicate key",
			Input:         "public: no\npublic: yes",
			ExpectedError: "duplicate key 'public'",
		},
		{
			Name:          "malformed key",
			Input:         "invalid: no",
			ExpectedError: "unrecognized key 'invalid'",
		},
		{
			Name:          "malformed value",
			Input:         "public: invalid",
			ExpectedError: "could not parse a line: value 'invalid' is not 'yes' or 'no'",
		},
		{
			Name:         "empty lines",
			Input:        "\n\n\npublic: yes\n\n",
			ResultPublic: true,
		},
		{
			Name:         "space around",
			Input:        "  public: yes  ",
			ResultPublic: true,
		},
		{
			Name:         "trailing newline",
			Input:        "public: yes\n",
			ResultPublic: true,
		},
		{
			Name:          "empty file",
			Input:         "",
			ExpectedError: "access file is empty",
		},
		{
			Name:          "newlines only",
			Input:         "\n\n\n\n\n",
			ExpectedError: "access file is empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			path, cleanup := fixture.File(t)
			defer cleanup()

			data := []byte(testCase.Input)
			writeToFile(t, path, data)

			l := library.NewDelimiterAccessLoader()

			access, err := l.Load(path)
			if testCase.ExpectedError != "" {
				require.EqualError(t, err, testCase.ExpectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ResultPublic, access.Public())
			}
		})
	}

}

func writeToFile(t *testing.T, path string, data []byte) {
	permissions := 0600 | os.ModePerm
	err := os.WriteFile(path, data, permissions)
	require.NoError(t, err)
}
