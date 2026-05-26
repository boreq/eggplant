package auth_test

import (
	"testing"

	"github.com/boreq/eggplant/adapters/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/stretchr/testify/require"
)

func TestPasswordHasher(t *testing.T) {
	password, err := authdomain.NewPasswordFromString("password")
	require.NoError(t, err)
	otherPassword, err := authdomain.NewPasswordFromString("other-password")
	require.NoError(t, err)

	h := auth.NewBcryptPasswordHasher()

	hash, err := h.Hash(password)
	require.NoError(t, err)

	err = h.Compare(hash, password)
	require.NoError(t, err)

	err = h.Compare(hash, otherPassword)
	require.Error(t, err)
}
