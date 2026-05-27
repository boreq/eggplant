package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/stretchr/testify/require"
)

var (
	adminCtx = auth.NewAdminAccessContext()
	anonCtx  = auth.NewAnonymousAccessContext()
)

func userFromCtx(t *testing.T, a *auth.Auth, ctx auth.AccessContext) authdomain.User {
	ac, ok := ctx.(auth.AuthenticatedAccessContext)
	require.True(t, ok)
	u, err := a.GetCurrentUser.Execute(ac)
	require.NoError(t, err)
	return u
}

func newAdminAuthCtx(t *testing.T, a *auth.Auth, username authdomain.Username, password authdomain.Password) auth.AuthenticatedAccessContext {
	err := a.RegisterInitial.Execute(auth.RegisterInitial{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	token, err := a.Login.Execute(anonCtx, auth.Login{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	ctx, err := a.CheckAccessToken.Execute(auth.CheckAccessToken{Token: token})
	require.NoError(t, err)

	authCtx, ok := ctx.(auth.AuthenticatedAccessContext)
	require.True(t, ok)
	return authCtx
}

func mustUsername(t *testing.T, s string) authdomain.Username {
	u, err := authdomain.NewUsernameFromString(s)
	require.NoError(t, err)
	return u
}

func mustPassword(t *testing.T, s string) authdomain.Password {
	p, err := authdomain.NewPasswordFromString(s)
	require.NoError(t, err)
	return p
}

func mustAccessToken(t *testing.T, s string) authdomain.AccessToken {
	tok, err := authdomain.NewAccessTokenFromString(s)
	require.NoError(t, err)
	return tok
}

func TestRegisterInitial(t *testing.T) {
	for _, testCase := range registerTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			a, cleanup := NewAuth(t)
			defer cleanup()

			username, usernameErr := authdomain.NewUsernameFromString(testCase.Username)
			password, passwordErr := authdomain.NewPasswordFromString(testCase.Password)

			if testCase.ExpectedError != nil {
				if usernameErr != nil {
					require.EqualError(t, usernameErr, testCase.ExpectedError.Error())
					return
				}
				if passwordErr != nil {
					require.EqualError(t, passwordErr, testCase.ExpectedError.Error())
					return
				}
			}
			require.NoError(t, usernameErr)
			require.NoError(t, passwordErr)

			cmd := auth.RegisterInitial{
				Username: username,
				Password: password,
			}

			err := a.RegisterInitial.Execute(cmd)
			require.NoError(t, err)

			users, err := a.List.Execute(adminCtx)
			require.NoError(t, err)

			require.Equal(t, 1, len(users))
			require.Equal(t, username, users[0].Username())
			require.Equal(t, true, users[0].Administrator())
			require.False(t, users[0].Created().IsZero())
			require.False(t, users[0].LastSeen().IsZero())
		})
	}
}

func TestRegisterInitialCanOnlyBePerformedOnce(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	cmd := auth.RegisterInitial{
		Username: mustUsername(t, "username"),
		Password: mustPassword(t, "password"),
	}

	err := a.RegisterInitial.Execute(cmd)
	require.NoError(t, err)

	err = a.RegisterInitial.Execute(cmd)
	require.EqualError(t, err, "transaction failed: there are existing users")
}

func TestLoginInitialUser(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	token, err := a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, token.String())

	_, err = a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: mustPassword(t, "other-password"),
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
	require.EqualError(t, err, "transaction failed: invalid password: unauthorized")

	_, err = a.Login.Execute(anonCtx,
		auth.Login{
			Username: mustUsername(t, "other-username"),
			Password: password,
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
	require.EqualError(t, err, "transaction failed: user not found: unauthorized")
}

func TestCheckAccessToken(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	err := a.RegisterInitial.Execute(auth.RegisterInitial{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	token, err := a.Login.Execute(anonCtx, auth.Login{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	ctx, err := a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: token},
	)
	require.NoError(t, err)

	u := userFromCtx(t, a, ctx)
	require.Equal(t, username, u.Username())
	require.Equal(t, true, u.Administrator())
	require.False(t, u.Created().IsZero())
	require.False(t, u.LastSeen().IsZero())

	_, err = a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: mustAccessToken(t, "fake")},
	)
	require.EqualError(t, err, "could not get the username: unauthorized")
	require.True(t, errors.Is(err, auth.ErrUnauthorized))

	_, err = a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: mustAccessToken(t, "fake-ab")},
	)
	require.EqualError(t, err, "transaction failed: user not found: unauthorized")
	require.True(t, errors.Is(err, auth.ErrUnauthorized))

	_, err = a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: mustAccessToken(t, "fake-757365726E616D65")},
	)
	require.EqualError(t, err, "transaction failed: invalid token: unauthorized")
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
}

func TestUpdateLastSeen(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	token, err := a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	ctx1, err := a.CheckAccessToken.Execute(
		auth.CheckAccessToken{
			Token: token,
		},
	)
	require.NoError(t, err)

	<-time.After(10 * time.Millisecond)

	ctx2, err := a.CheckAccessToken.Execute(
		auth.CheckAccessToken{
			Token: token,
		},
	)
	require.NoError(t, err)

	u1 := userFromCtx(t, a, ctx1)
	u2 := userFromCtx(t, a, ctx2)
	require.False(t, u1.Created().IsZero())
	require.False(t, u1.LastSeen().IsZero())
	require.False(t, u2.Created().IsZero())
	require.False(t, u2.LastSeen().IsZero())
	require.Equal(t, u1.Created(), u2.Created())
}

func TestLogout(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	token, err := a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	ctx, err := a.CheckAccessToken.Execute(auth.CheckAccessToken{Token: token})
	require.NoError(t, err)

	authCtx, ok := ctx.(auth.AuthenticatedAccessContext)
	require.True(t, ok)

	err = a.Logout.Execute(authCtx)
	require.NoError(t, err)
}

func TestList(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	users, err := a.List.Execute(adminCtx)
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, username, users[0].Username())
	require.Equal(t, true, users[0].Administrator())
	require.False(t, users[0].Created().IsZero())
	require.False(t, users[0].LastSeen().IsZero())
}

func TestCreateInvitation(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	token, err := a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)
	require.NotEmpty(t, token.String())
}

func TestRegisterInvalidInvitationToken(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	invalidToken, err := authdomain.NewInvitationTokenFromString("invalid")
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: mustUsername(t, "username"),
			Password: mustPassword(t, "password"),
			Token:    invalidToken,
		},
	)
	require.Error(t, err)
}

func TestRegisterTokenCanNotBeReused(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	token, err := a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: mustUsername(t, "username"),
			Password: mustPassword(t, "password"),
			Token:    token,
		},
	)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: mustUsername(t, "other-username"),
			Password: mustPassword(t, "other-password"),
			Token:    token,
		},
	)
	require.Error(t, err)
	require.EqualError(t, err, "transaction failed: could not get the invitation: not found")
}

func TestRegisterUsernameCanNotBeTaken(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	token, err := a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: username,
			Password: password,
			Token:    token,
		},
	)
	require.NoError(t, err)

	token, err = a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: username,
			Password: password,
			Token:    token,
		},
	)
	require.EqualError(t, err, "transaction failed: username taken")
	require.True(t, errors.Is(err, auth.ErrUsernameTaken))
}

func TestRegisterInvalid(t *testing.T) {
	for _, testCase := range registerTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			a, cleanup := NewAuth(t)
			defer cleanup()

			username, usernameErr := authdomain.NewUsernameFromString(testCase.Username)
			password, passwordErr := authdomain.NewPasswordFromString(testCase.Password)

			if testCase.ExpectedError != nil {
				if usernameErr != nil {
					require.EqualError(t, usernameErr, testCase.ExpectedError.Error())
					return
				}
				if passwordErr != nil {
					require.EqualError(t, passwordErr, testCase.ExpectedError.Error())
					return
				}
			}
			require.NoError(t, usernameErr)
			require.NoError(t, passwordErr)

			token, err := a.CreateInvitation.Execute(adminCtx)
			require.NoError(t, err)

			err = a.Register.Execute(anonCtx,
				auth.Register{
					Username: username,
					Password: password,
					Token:    token,
				},
			)
			require.NoError(t, err)

			users, err := a.List.Execute(adminCtx)
			require.NoError(t, err)
			require.Equal(t, 1, len(users))
			require.Equal(t, username, users[0].Username())
			require.Equal(t, false, users[0].Administrator())
			require.False(t, users[0].Created().IsZero())
			require.False(t, users[0].LastSeen().IsZero())
		})
	}
}

func TestLogin(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	invitationToken, err := a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: username,
			Password: password,
			Token:    invitationToken,
		},
	)
	require.NoError(t, err)

	accessToken, err := a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken.String())

	_, err = a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: mustPassword(t, "other-password"),
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))

	_, err = a.Login.Execute(anonCtx,
		auth.Login{
			Username: mustUsername(t, "other-username"),
			Password: password,
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
}

func TestRemove(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	adminAuthCtx := newAdminAuthCtx(t, a, mustUsername(t, "admin"), mustPassword(t, "password"))

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")

	invitationToken, err := a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: username,
			Password: password,
			Token:    invitationToken,
		},
	)
	require.NoError(t, err)

	users, err := a.List.Execute(adminCtx)
	require.NoError(t, err)
	require.Equal(t, 2, len(users))

	err = a.Remove.Execute(adminAuthCtx,
		auth.Remove{
			Username: username,
		},
	)
	require.NoError(t, err)

	users, err = a.List.Execute(adminCtx)
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
}

func TestRemoveNoUser(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	adminAuthCtx := newAdminAuthCtx(t, a, mustUsername(t, "admin"), mustPassword(t, "password"))

	users, err := a.List.Execute(adminCtx)
	require.NoError(t, err)
	require.Equal(t, 1, len(users))

	err = a.Remove.Execute(adminAuthCtx,
		auth.Remove{
			Username: mustUsername(t, "username"),
		},
	)
	require.NoError(t, err)

	users, err = a.List.Execute(adminCtx)
	require.NoError(t, err)

	require.Equal(t, 1, len(users))
}

func TestSetPassword(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	username := mustUsername(t, "username")
	password := mustPassword(t, "password")
	newPassword := mustPassword(t, "new-password")

	invitationToken, err := a.CreateInvitation.Execute(adminCtx)
	require.NoError(t, err)

	err = a.Register.Execute(anonCtx,
		auth.Register{
			Username: username,
			Password: password,
			Token:    invitationToken,
		},
	)
	require.NoError(t, err)

	_, err = a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	err = a.SetPassword.Execute(adminCtx,
		auth.SetPassword{
			Username: username,
			Password: newPassword,
		},
	)
	require.NoError(t, err)

	_, err = a.Login.Execute(anonCtx,
		auth.Login{
			Username: username,
			Password: newPassword,
		},
	)
	require.NoError(t, err)
}

func NewAuth(t *testing.T) (*auth.Auth, fixture.CleanupFunc) {
	db, cleanup := fixture.Bolt(t)

	a, err := wire.BuildAuthForTest(db)
	if err != nil {
		t.Fatal(err)
	}

	return a, cleanup
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
		ExpectedError: errors.New("username can't be empty"),
	},
	{
		Name:          "empty_password",
		Username:      "username",
		Password:      "",
		ExpectedError: errors.New("password can't be empty"),
	},
	{
		Name:          "username_too_long",
		Username:      strings.Repeat("a", 101),
		Password:      "password",
		ExpectedError: errors.New("username length can't exceed 100 characters"),
	},
	{
		Name:          "password_too_long",
		Username:      "username",
		Password:      strings.Repeat("a", 10001),
		ExpectedError: errors.New("password length can't exceed 10000 characters"),
	},
}
