package commands

import (
	"runtime"
	"time"

	"github.com/boreq/plum/config"
	"github.com/boreq/plum/core"
	"github.com/boreq/plum/parser"
	"github.com/boreq/plum/server"
	"github.com/boreq/guinea"
	"github.com/dustin/go-humanize"
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

	r := core.NewRepository(conf)

	tracker := core.NewTracker(p, r)
	errC := make(chan error)

	// Statistics
	go func() {
		lastLines, _ := tracker.GetStats()
		duration := 1 * time.Second
		for range time.Tick(duration) {
			lines, _ := tracker.GetStats()
			linesPerSecond := float64(lines-lastLines) / duration.Seconds()
			log.Debug("data statistics", "totalLines", lines, "linesPerSecond", linesPerSecond)
			lastLines = lines
			logMemoryStats()
		}
	}()

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

func logMemoryStats() {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	alloc := humanize.Bytes(m.Alloc)
	totalAlloc := humanize.Bytes(m.TotalAlloc)
	sys := humanize.Bytes(m.Sys)
	numGC := m.NumGC

	log.Debug("memory statistics", "alloc", alloc, "totalAlloc", totalAlloc, "sys", sys, "numGC", numGC)
}
