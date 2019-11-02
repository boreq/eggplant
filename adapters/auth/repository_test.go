package auth_test

import (
	"errors"
	"testing"

	"github.com/boreq/eggplant/adapters/auth"
	appAuth "github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
)

func TestRegisterInitial(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial(username, password)
	require.NoError(t, err)

	err = r.RegisterInitial(username, password)
	require.Error(t, err)
}

func TestLogin(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial(username, password)
	require.NoError(t, err)

	token, err := r.Login(username, password)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	_, err = r.Login(username, "other-password")
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))

	_, err = r.Login("other-username", password)
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))
}

func TestCheckAccessToken(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial("username", "password")
	require.NoError(t, err)

	// checking a real token should work
	token, err := r.Login(username, password)
	require.NoError(t, err)

	u, err := r.CheckAccessToken(token)
	require.NoError(t, err)

	require.Equal(t, u.Username, username)
	require.Equal(t, u.Administrator, true)

	// checking a made up token should fail
	_, err = r.CheckAccessToken(appAuth.AccessToken("fake"))
	require.EqualError(t, err, "could not get the username: unauthorized")
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))

	_, err = r.CheckAccessToken(appAuth.AccessToken("fake-ab"))
	require.EqualError(t, err, "transaction failed: invalid username: unauthorized")
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))

	_, err = r.CheckAccessToken(appAuth.AccessToken("fake-757365726E616D65"))
	require.EqualError(t, err, "transaction failed: invalid token: unauthorized")
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))
}

func TestLogout(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial("username", "password")
	require.NoError(t, err)

	token, err := r.Login(username, password)
	require.NoError(t, err)

	err = r.Logout(token)
	require.NoError(t, err)

	err = r.Logout(appAuth.AccessToken("fake"))
	require.EqualError(t, err, "could not extract the username: malformed token")

	err = r.Logout(appAuth.AccessToken("fake-ab"))
	require.EqualError(t, err, "transaction failed: user doesn't exist")

	err = r.Logout(appAuth.AccessToken("fake-757365726E616D65"))
	require.EqualError(t, err, "transaction failed: session not found")
}

func TestCount(t *testing.T) {
	r, cleanup := NewRepository(t)
	defer cleanup()

	n, err := r.Count()
	require.NoError(t, err)
	require.Equal(t, 0, n)

	err = r.RegisterInitial("username", "password")
	require.NoError(t, err)

	n, err = r.Count()
	require.NoError(t, err)
	require.Equal(t, 1, n)
}

func TestList(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial(username, password)
	require.NoError(t, err)

	users, err := r.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, username, users[0].Username)
	require.Equal(t, true, users[0].Administrator)
}

func TestCreateInvitation(t *testing.T) {
	r, cleanup := NewRepository(t)
	defer cleanup()

	token, err := r.CreateInvitation()
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestRegisterInvalid(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.Register(username, password, appAuth.InvitationToken(""))
	require.Error(t, err)
}

func TestRegisterTokenCanNotBeReused(t *testing.T) {
	r, cleanup := NewRepository(t)
	defer cleanup()

	token, err := r.CreateInvitation()
	require.NoError(t, err)

	err = r.Register("username", "password", token)
	require.NoError(t, err)

	err = r.Register("other-username", "other-password", token)
	require.Error(t, err)
	require.EqualError(t, err, "transaction failed: token does not exist")
}

func TestRegisterUsernameCanNotBeTaken(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	token, err := r.CreateInvitation()
	require.NoError(t, err)

	err = r.Register(username, password, token)
	require.NoError(t, err)

	users, err := r.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, username, users[0].Username)
	require.Equal(t, false, users[0].Administrator)

	token, err = r.CreateInvitation()
	require.NoError(t, err)

	err = r.Register(username, password, token)
	require.Error(t, err)
	require.True(t, errors.Is(err, appAuth.ErrUsernameTaken))
	require.EqualError(t, err, "transaction failed: username taken")
}

func NewRepository(t *testing.T) (*auth.UserRepository, fixture.CleanupFunc) {
	db, cleanup := fixture.Bolt(t)

	ph := auth.NewBcryptPasswordHasher()
	atg := auth.NewCryptoAccessTokenGenerator()

	r, err := auth.NewUserRepository(db, ph, atg)
	if err != nil {
		t.Fatal(err)
	}

	return r, cleanup

}
