package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/boreq/eggplant/adapters/auth"
	appAuth "github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/stretchr/testify/require"
)

func TestRegisterInitial(t *testing.T) {
	for _, testCase := range registerTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r, cleanup := NewRepository(t)
			defer cleanup()

			err := r.RegisterInitial(testCase.Username, testCase.Password)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)

				users, err := r.List()
				require.NoError(t, err)

				require.Equal(t, 1, len(users))
				require.Equal(t, testCase.Username, users[0].Username)
				require.Equal(t, true, users[0].Administrator)
				require.False(t, users[0].Created.IsZero())
				require.False(t, users[0].LastSeen.IsZero())
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestRegisterInitialCanOnlyBePerformedOnce(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial(username, password)
	require.NoError(t, err)

	err = r.RegisterInitial(username, password)
	require.EqualError(t, err, "there are existing users")
}

func TestLoginInitialUser(t *testing.T) {
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

	require.Equal(t, username, u.Username)
	require.Equal(t, true, u.Administrator)
	require.False(t, u.Created.IsZero())
	require.False(t, u.LastSeen.IsZero())

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

func TestUpdateLastSeen(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.RegisterInitial("username", "password")
	require.NoError(t, err)

	token, err := r.Login(username, password)
	require.NoError(t, err)

	u1, err := r.CheckAccessToken(token)
	require.NoError(t, err)

	<-time.After(10 * time.Millisecond)

	u2, err := r.CheckAccessToken(token)
	require.NoError(t, err)

	require.False(t, u1.Created.IsZero())
	require.False(t, u1.LastSeen.IsZero())
	require.False(t, u2.Created.IsZero())
	require.False(t, u2.LastSeen.IsZero())
	require.Equal(t, u1.Created, u2.Created)
	require.NotEqual(t, u1.LastSeen, u2.LastSeen)
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
	require.False(t, users[0].Created.IsZero())
	require.False(t, users[0].LastSeen.IsZero())
}

func TestCreateInvitation(t *testing.T) {
	r, cleanup := NewRepository(t)
	defer cleanup()

	token, err := r.CreateInvitation()
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestRegisterInvalidInvitationToken(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	err := r.Register(username, password, appAuth.InvitationToken("invalid"))
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

	token, err = r.CreateInvitation()
	require.NoError(t, err)

	err = r.Register(username, password, token)
	require.EqualError(t, err, "transaction failed: username taken")
	require.True(t, errors.Is(err, appAuth.ErrUsernameTaken))
}

func TestRegisterInvalid(t *testing.T) {
	for _, testCase := range registerTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r, cleanup := NewRepository(t)
			defer cleanup()

			token, err := r.CreateInvitation()
			require.NoError(t, err)

			err = r.Register(testCase.Username, testCase.Password, token)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)

				users, err := r.List()
				require.NoError(t, err)
				require.Equal(t, 1, len(users))
				require.Equal(t, testCase.Username, users[0].Username)
				require.Equal(t, false, users[0].Administrator)
				require.False(t, users[0].Created.IsZero())
				require.False(t, users[0].LastSeen.IsZero())
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestLogin(t *testing.T) {
	const username = "username"
	const password = "password"

	r, cleanup := NewRepository(t)
	defer cleanup()

	invitationToken, err := r.CreateInvitation()
	require.NoError(t, err)

	err = r.Register(username, password, invitationToken)
	require.NoError(t, err)

	accessToken, err := r.Login(username, password)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)

	_, err = r.Login(username, "other-password")
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))

	_, err = r.Login("other-username", password)
	require.True(t, errors.Is(err, appAuth.ErrUnauthorized))
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

var registerTestCases = []struct {
	Name string

	Username string
	Password string

	ExpectedError error
}{
	{
		Name:          "valid",
		Username:      "username",
		Password:      "password",
		ExpectedError: nil,
	},
	{
		Name:          "empty_username",
		Username:      "",
		Password:      "password",
		ExpectedError: errors.New("invalid parameters: username can't be empty"),
	},
	{
		Name:          "empty_password",
		Username:      "username",
		Password:      "",
		ExpectedError: errors.New("invalid parameters: password can't be empty"),
	},
	{
		Name:          "username_too_long",
		Username:      strings.Repeat("a", 101),
		Password:      "password",
		ExpectedError: errors.New("invalid parameters: username length can't exceed 100 characters"),
	},
	{
		Name:          "password_too_long",
		Username:      "username",
		Password:      strings.Repeat("a", 10001),
		ExpectedError: errors.New("invalid parameters: password length can't exceed 10000 characters"),
	},
}
