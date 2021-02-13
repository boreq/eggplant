package config_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/boreq/eggplant/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGolden(t *testing.T) {
	conf := config.Default()

	actual := &bytes.Buffer{}

	err := config.Marshal(actual, conf.ExposedConfig)
	require.NoError(t, err)

	f, err := os.Open("data/config.golden.toml")
	require.NoError(t, err)

	defer f.Close()

	expected, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	require.Equal(t, string(expected), string(actual.Bytes()))
}
