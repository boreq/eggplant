package remote_test

import (
	"testing"

	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/stretchr/testify/require"
)

func TestNewRemoteInstanceName(t *testing.T) {
	testCases := []struct {
		Address      string
		ExpectedName string
	}{
		{
			Address:      "http://example.com",
			ExpectedName: "example.com",
		},
		{
			Address:      "https://example.com",
			ExpectedName: "example.com",
		},
		{
			Address:      "http://example.com/",
			ExpectedName: "example.com",
		},
		{
			Address:      "http://example.com/whatever",
			ExpectedName: "example.com/whatever",
		},
		{
			Address:      "http://example.com/whatever/",
			ExpectedName: "example.com/whatever",
		},
		{
			Address:      "http://example.com:1234/whatever",
			ExpectedName: "example.com:1234/whatever",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Address, func(t *testing.T) {
			address, err := remotedomain.NewRemoteInstanceAddress(testCase.Address)
			require.NoError(t, err)

			name, err := remotedomain.NewRemoteInstanceName(address)
			require.NoError(t, err)
			require.Equal(t, testCase.ExpectedName, name.String())
		})
	}
}
