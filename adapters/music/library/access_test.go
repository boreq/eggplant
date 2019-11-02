package library_test

import (
	"io/ioutil"
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
		ExpectedError bool
		ResultPublic  bool
	}{
		{
			Name:          "yes",
			Input:         "public: yes",
			ExpectedError: false,
			ResultPublic:  true,
		},
		{
			Name:          "no",
			Input:         "public: no",
			ExpectedError: false,
			ResultPublic:  false,
		},
		{
			Name:          "overwrite",
			Input:         "public: no\npublic: yes",
			ExpectedError: false,
			ResultPublic:  true,
		},
		{
			Name:          "malformed key",
			Input:         "invalid: no",
			ExpectedError: true,
		},
		{
			Name:          "malformed value",
			Input:         "public: invalid",
			ExpectedError: true,
		},
		{
			Name:          "empty lines",
			Input:         "public: no\n\n\npublic:yes",
			ExpectedError: false,
			ResultPublic:  true,
		},
		{
			Name:          "space around",
			Input:         "  public: no\n\n\n public:yes  ",
			ExpectedError: false,
			ResultPublic:  true,
		},
		{
			Name:          "trailing newline",
			Input:         "public: yes\n",
			ExpectedError: false,
			ResultPublic:  true,
		},
		{
			Name:          "empty file",
			Input:         "",
			ExpectedError: true,
		},
		{
			Name:          "newlines only",
			Input:         "\n\n\n\n\n",
			ExpectedError: true,
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
			if testCase.ExpectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ResultPublic, access.Public)
			}
		})
	}

}

func writeToFile(t *testing.T, path string, data []byte) {
	permissions := 0600 | os.ModePerm
	err := ioutil.WriteFile(path, data, permissions)
	require.NoError(t, err)
}
