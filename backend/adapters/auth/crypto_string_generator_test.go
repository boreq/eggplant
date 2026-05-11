package auth_test

import (
	"testing"

	"github.com/boreq/eggplant/adapters/auth"
	"github.com/stretchr/testify/require"
)

func TestCryptoStringGenerator(t *testing.T) {
	g := auth.NewCryptoStringGenerator()
	s, err := g.Generate(10)
	require.NoError(t, err)
	require.NotEmpty(t, s)
}
