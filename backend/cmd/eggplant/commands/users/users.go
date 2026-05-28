package users

import (
	"encoding/json"
	"fmt"

	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/eggplant/internal/wire"
	"github.com/boreq/guinea"
	"github.com/pkg/errors"
)

var UsersCmd = guinea.Command{
	Run: runUsers,
	Subcommands: map[string]*guinea.Command{
		"list":           &listCmd,
		"reset_password": &resetPasswordCmd,
	},
	ShortDescription: "manage users",
}

func runUsers(c guinea.Context) error {
	return guinea.ErrInvalidParms
}

var listCmd = guinea.Command{
	Run: runList,
	Arguments: []guinea.Argument{
		{
			Name:        "data_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to the directory used for data storage",
		},
	},
	ShortDescription: "list all users",
}

func runList(c guinea.Context) error {
	conf := config.Default()
	conf.DataDirectory = c.Arguments[0]

	a, err := wire.BuildAuth(conf)
	if err != nil {
		return errors.Wrap(err, "failed to build the application")
	}

	users, err := a.List.Execute(auth.NewCommandLineAccessContext())
	if err != nil {
		return errors.Wrap(err, "failed to list users")
	}

	j, err := json.MarshalIndent(toUserList(users), "", "    ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal to json")
	}

	fmt.Println(string(j))

	return nil
}

type userOutput struct {
	Username      string `json:"username"`
	Administrator bool   `json:"administrator"`
}

func toUserList(users []authdomain.User) []userOutput {
	rv := make([]userOutput, 0, len(users))
	for _, u := range users {
		rv = append(rv, userOutput{
			Username:      u.Username().String(),
			Administrator: u.Administrator(),
		})
	}
	return rv
}

var resetPasswordCmd = guinea.Command{
	Run: runResetPassword,
	Arguments: []guinea.Argument{
		{
			Name:        "data_directory",
			Optional:    false,
			Multiple:    false,
			Description: "Path to the directory used for data storage",
		},
		{
			Name:        "username",
			Optional:    false,
			Multiple:    false,
			Description: "Username",
		},
	},
	ShortDescription: "resets a user's password",
}

func runResetPassword(c guinea.Context) error {
	conf := config.Default()
	conf.DataDirectory = c.Arguments[0]

	a, err := wire.BuildAuth(conf)
	if err != nil {
		return errors.Wrap(err, "failed to build the application")
	}

	username, err := authdomain.NewUsernameFromString(c.Arguments[1])
	if err != nil {
		return errors.Wrap(err, "invalid username")
	}

	rawPassword, err := authdomain.GenerateRandomPassword()
	if err != nil {
		return errors.Wrap(err, "failed to generate a password")
	}

	cmd := auth.SetPassword{
		Username: username,
		Password: rawPassword,
	}

	if err := a.SetPassword.Execute(auth.NewCommandLineAccessContext(), cmd); err != nil {
		return errors.Wrap(err, "failed to set a password")
	}

	out := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username.String(),
		Password: rawPassword.String(),
	}

	j, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal to json")
	}

	fmt.Println(string(j))

	return nil
}
