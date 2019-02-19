package commands

import (
	"github.com/boreq/goaccess/config"
	"github.com/boreq/goaccess/core"
	"github.com/boreq/goaccess/parser"
	"github.com/boreq/goaccess/server"
	"github.com/boreq/guinea"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "follow",
			Optional:    false,
			Multiple:    false,
			Description: "a log file to be monitored",
		},
		{
			Name:        "load",
			Optional:    true,
			Multiple:    true,
			Description: "log files to be initially loaded",
		},
	},
	ShortDescription: "runs the program",
}

func runRun(c guinea.Context) error {
	conf := config.Default()
	//if err := config.Load(c.Arguments[0]); err != nil {
	//	return err
	//}
	//m := monitor.New(config.Config.ScriptsDirectory, config.Config.UpdateEverySeconds)

	//if err := server.Serve(m, config.Config.ServeAddress); err != nil {
	//	return err
	//}
	comb := parser.PredefinedFormats["combined"]
	p, err := parser.NewParser(comb)
	if err != nil {
		return err
	}

	tracker := core.NewTracker(p)
	errC := make(chan error)

	// Load the specified files.
	for i := 1; i < len(c.Arguments); i++ {
		go func(i int) {
			errC <- tracker.Load(c.Arguments[i])
		}(i)
	}

	for i := 1; i < len(c.Arguments); i++ {
		err := <-errC
		if err != nil {
			return err
		}
	}

	// Track the specified file.
	go func() {
		errC <- tracker.Follow(c.Arguments[0])
	}()

	go func() {
		errC <- server.Serve(tracker.Repository, conf.ServeAddress)
	}()

	return <-errC
}
