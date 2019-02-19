package commands

import (
	"github.com/boreq/goaccess/logging"
	"github.com/boreq/guinea"
)

var log = logging.New("main/commands")

var MainCmd = guinea.Command{
	Run: runMain,
	Subcommands: map[string]*guinea.Command{
		"run": &runCmd,
	},
	ShortDescription: "live log analyzer",
	Description:      "This software analyzes web server logs in real time.",
}

func runMain(c guinea.Context) error {
	return guinea.ErrInvalidParms
}
