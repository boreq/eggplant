package commands

import (
	"github.com/boreq/eggplant/cmd/eggplant/commands/users"
	"github.com/boreq/guinea"
)

var MainCmd = guinea.Command{
	Run: runMain,
	Subcommands: map[string]*guinea.Command{
		"run":            &runCmd,
		"default_config": &defaultConfigCmd,
		"users":          &users.UsersCmd,
	},
	ShortDescription: "a music streaming service",
	Description: `
Eggplant serves your music using a web interface.
`,
}

func runMain(c guinea.Context) error {
	return guinea.ErrInvalidParms
}
