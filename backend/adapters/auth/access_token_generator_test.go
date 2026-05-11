package auth_test

import (
	"testing"

	"github.com/boreq/eggplant/adapters/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	const username = "username"

	g := auth.NewCryptoAccessTokenGenerator()

	token, err := g.Generate(username)
	require.NoError(t, err)

	retrievedUsername, err := g.GetUsername(token)
	require.NoError(t, err)
	assert.Equal(t, username, retrievedUsername)
}

func TestMalformed(t *testing.T) {
	g := auth.NewCryptoAccessTokenGenerator()
	_, err := g.GetUsername("invalid")
	require.EqualError(t, err, "malformed token")
}
