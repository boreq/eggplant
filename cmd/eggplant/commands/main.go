package commands

import (
	"github.com/boreq/guinea"
)

var MainCmd = guinea.Command{
	Run: runMain,
	Subcommands: map[string]*guinea.Command{
		"run": &runCmd,
	},
	ShortDescription: "a music streaming service",
	Description: `
Eggplant serves your music using a web interface.
`,
}

func runMain(c guinea.Context) error {
	return guinea.ErrInvalidParms
}
