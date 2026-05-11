package auth_test

import (
	"testing"

	"github.com/boreq/eggplant/adapters/auth"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher(t *testing.T) {
	const password = "password"
	const otherPassword = "other-password"

	h := auth.NewBcryptPasswordHasher()

	hash, err := h.Hash(password)
	require.NoError(t, err)

	err = h.Compare(hash, password)
	require.NoError(t, err)

	err = h.Compare(hash, otherPassword)
	require.Error(t, err)
}
