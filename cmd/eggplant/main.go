// Package main contains the main function.
package main

import (
	"fmt"
	"os"

	"github.com/boreq/eggplant/cmd/eggplant/commands"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/guinea"
)

func main() {
	injectGlobalBehaviour(&commands.MainCmd)
	if err := guinea.Run(&commands.MainCmd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var globalOptions = []guinea.Option{
	guinea.Option{
		Name:        "verbosity",
		Type:        guinea.String,
		Default:     "info",
		Description: "One of: debug, info, warn, error or crit. Default: info",
	},
}

func injectGlobalBehaviour(cmd *guinea.Command) {
	cmd.Options = append(cmd.Options, globalOptions...)
	oldRun := cmd.Run
	cmd.Run = func(c guinea.Context) error {
		level, err := logging.LevelFromString(c.Options["verbosity"].Str())
		if err != nil {
			return err
		}
		logging.SetLoggingLevel(level)
		if oldRun != nil {
			return oldRun(c)
		}
		return nil
	}
	for _, subCmd := range cmd.Subcommands {
		injectGlobalBehaviour(subCmd)
	}
}
