package tests

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/internal/fixture"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/stretchr/testify/require"
)

func TestRegisterInitial(t *testing.T) {
	for _, testCase := range registerTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			a, cleanup := NewAuth(t)
			defer cleanup()

			cmd := auth.RegisterInitial{
				Username: testCase.Username,
				Password: testCase.Password,
			}

			err := a.RegisterInitial.Execute(cmd)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)

				users, err := a.List.Execute()
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
	a, cleanup := NewAuth(t)
	defer cleanup()

	cmd := auth.RegisterInitial{
		Username: "username",
		Password: "password",
	}

	err := a.RegisterInitial.Execute(cmd)
	require.NoError(t, err)

	err = a.RegisterInitial.Execute(cmd)
	require.EqualError(t, err, "transaction failed: there are existing users")
}

func TestLoginInitialUser(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	token, err := a.Login.Execute(
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	_, err = a.Login.Execute(
		auth.Login{
			Username: username,
			Password: "other-password",
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
	require.EqualError(t, err, "transaction failed: invalid password: unauthorized")

	_, err = a.Login.Execute(
		auth.Login{
			Username: "other-username",
			Password: password,
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
	require.EqualError(t, err, "transaction failed: user not found: unauthorized")
}

func TestCheckAccessToken(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	err := a.RegisterInitial.Execute(auth.RegisterInitial{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	// checking a real token should work
	token, err := a.Login.Execute(auth.Login{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	u, err := a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: token},
	)
	require.NoError(t, err)

	require.Equal(t, username, u.Username)
	require.Equal(t, true, u.Administrator)
	require.False(t, u.Created.IsZero())
	require.False(t, u.LastSeen.IsZero())

	// checking a made up token should fail
	_, err = a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: "fake"},
	)
	require.EqualError(t, err, "could not get the username: unauthorized")
	require.True(t, errors.Is(err, auth.ErrUnauthorized))

	_, err = a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: "fake-ab"},
	)
	require.EqualError(t, err, "transaction failed: user not found: unauthorized")
	require.True(t, errors.Is(err, auth.ErrUnauthorized))

	_, err = a.CheckAccessToken.Execute(
		auth.CheckAccessToken{Token: "fake-757365726E616D65"},
	)
	require.EqualError(t, err, "transaction failed: invalid token: unauthorized")
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
}

func TestUpdateLastSeen(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	token, err := a.Login.Execute(
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	u1, err := a.CheckAccessToken.Execute(
		auth.CheckAccessToken{
			Token: token,
		},
	)
	require.NoError(t, err)

	<-time.After(10 * time.Millisecond)

	u2, err := a.CheckAccessToken.Execute(
		auth.CheckAccessToken{
			Token: token,
		},
	)
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

	a, cleanup := NewAuth(t)
	defer cleanup()

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	token, err := a.Login.Execute(
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	err = a.Logout.Execute(auth.Logout{Token: token})
	require.NoError(t, err)

	err = a.Logout.Execute(auth.Logout{Token: "fake"})
	require.EqualError(t, err, "could not extract the username: malformed token")

	err = a.Logout.Execute(auth.Logout{Token: "fake-ab"})
	require.EqualError(t, err, "transaction failed: could not get the user: not found")

	err = a.Logout.Execute(auth.Logout{Token: "fake-757365726E616D65"})
	require.EqualError(t, err, "transaction failed: session not found")
}

//func TestCount(t *testing.T) {
//	a, cleanup := NewAuth(t)
//	defer cleanup()
//
//	n, err := a.List.Execute()
//	require.NoError(t, err)
//	require.Equal(t, 0, n)
//
//	err = r.RegisterInitial("username", "password")
//	require.NoError(t, err)
//
//	n, err = r.Count()
//	require.NoError(t, err)
//	require.Equal(t, 1, n)
//}

func TestList(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	err := a.RegisterInitial.Execute(
		auth.RegisterInitial{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	users, err := a.List.Execute()
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, username, users[0].Username)
	require.Equal(t, true, users[0].Administrator)
	require.False(t, users[0].Created.IsZero())
	require.False(t, users[0].LastSeen.IsZero())
}

func TestCreateInvitation(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	token, err := a.CreateInvitation.Execute()
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestRegisterInvalidInvitationToken(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	err := a.Register.Execute(
		auth.Register{
			Username: username,
			Password: password,
			Token:    auth.InvitationToken("invalid"),
		},
	)
	require.Error(t, err)
}

func TestRegisterTokenCanNotBeReused(t *testing.T) {
	a, cleanup := NewAuth(t)
	defer cleanup()

	token, err := a.CreateInvitation.Execute()
	require.NoError(t, err)

	err = a.Register.Execute(
		auth.Register{
			Username: "username",
			Password: "password",
			Token:    token,
		},
	)
	require.NoError(t, err)

	err = a.Register.Execute(
		auth.Register{
			Username: "other-username",
			Password: "other-password",
			Token:    token,
		},
	)
	require.Error(t, err)
	require.EqualError(t, err, "transaction failed: could not get the invitation: not found")
}

func TestRegisterUsernameCanNotBeTaken(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	token, err := a.CreateInvitation.Execute()
	require.NoError(t, err)

	err = a.Register.Execute(
		auth.Register{
			Username: username,
			Password: password,
			Token:    token,
		},
	)
	require.NoError(t, err)

	token, err = a.CreateInvitation.Execute()
	require.NoError(t, err)

	err = a.Register.Execute(
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

			token, err := a.CreateInvitation.Execute()
			require.NoError(t, err)

			err = a.Register.Execute(
				auth.Register{
					Username: testCase.Username,
					Password: testCase.Password,
					Token:    token,
				},
			)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)

				users, err := a.List.Execute()
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

	a, cleanup := NewAuth(t)
	defer cleanup()

	invitationToken, err := a.CreateInvitation.Execute()
	require.NoError(t, err)

	err = a.Register.Execute(
		auth.Register{
			Username: username,
			Password: password,
			Token:    invitationToken,
		},
	)
	require.NoError(t, err)

	accessToken, err := a.Login.Execute(
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)

	_, err = a.Login.Execute(
		auth.Login{
			Username: username,
			Password: "other-password",
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))

	_, err = a.Login.Execute(
		auth.Login{
			Username: "other-username",
			Password: password,
		},
	)
	require.True(t, errors.Is(err, auth.ErrUnauthorized))
}

func TestRemove(t *testing.T) {
	const username = "username"
	const password = "password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	invitationToken, err := a.CreateInvitation.Execute()
	require.NoError(t, err)

	err = a.Register.Execute(
		auth.Register{
			Username: username,
			Password: password,
			Token:    invitationToken,
		},
	)
	require.NoError(t, err)

	users, err := a.List.Execute()
	require.NoError(t, err)
	require.Equal(t, 1, len(users))

	err = a.Remove.Execute(
		auth.Remove{
			Username: username,
		},
	)
	require.NoError(t, err)

	users, err = a.List.Execute()
	require.NoError(t, err)
	require.Equal(t, 0, len(users))
}

func TestRemoveNoUser(t *testing.T) {
	const username = "username"

	a, cleanup := NewAuth(t)
	defer cleanup()

	users, err := a.List.Execute()
	require.NoError(t, err)
	require.Equal(t, 0, len(users))

	err = a.Remove.Execute(
		auth.Remove{
			Username: username,
		},
	)
	require.NoError(t, err)

	users, err = a.List.Execute()
	require.NoError(t, err)

	require.Equal(t, 0, len(users))
}

func TestSetPassword(t *testing.T) {
	const username = "username"
	const password = "password"
	const newPassword = "new-password"

	a, cleanup := NewAuth(t)
	defer cleanup()

	invitationToken, err := a.CreateInvitation.Execute()
	require.NoError(t, err)

	err = a.Register.Execute(
		auth.Register{
			Username: username,
			Password: password,
			Token:    invitationToken,
		},
	)
	require.NoError(t, err)

	_, err = a.Login.Execute(
		auth.Login{
			Username: username,
			Password: password,
		},
	)
	require.NoError(t, err)

	err = a.SetPassword.Execute(
		auth.SetPassword{
			Username: username,
			Password: newPassword,
		},
	)
	require.NoError(t, err)

	_, err = a.Login.Execute(
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
