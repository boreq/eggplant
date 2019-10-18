package commands

import (
	"github.com/boreq/guinea"
)

var MainCmd = guinea.Command{
	Run: runMain,
	Subcommands: map[string]*guinea.Command{
		"run": &runCmd,
	},
	ShortDescription: "a real-time access log analyser",
	Description: `
Plum analyses web server access logs in real time and allows the user to access
the produced statistics using a web dashboard.
`,
}

func runMain(c guinea.Context) error {
	return guinea.ErrInvalidParms
}
