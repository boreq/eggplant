package commands

import (
	"runtime"
	"time"

	"github.com/boreq/guinea"
	"github.com/boreq/plum/config"
	"github.com/boreq/plum/core"
	"github.com/boreq/plum/parser"
	"github.com/boreq/plum/server"
	"github.com/dustin/go-humanize"
)

var runCmd = guinea.Command{
	Run: runRun,
	Arguments: []guinea.Argument{
		{
			Name:        "follow",
			Optional:    false,
			Multiple:    false,
			Description: "Log file to be monitored",
		},
		{
			Name:        "load",
			Optional:    true,
			Multiple:    true,
			Description: "Log files to be initially loaded",
		},
	},
	Options: []guinea.Option{
		guinea.Option{
			Name:        "address",
			Type:        guinea.String,
			Description: "Serve address",
			Default:     config.Default().ServeAddress,
		},
		guinea.Option{
			Name:        "log-format",
			Type:        guinea.String,
			Description: "Log format",
			Default:     config.Default().LogFormat,
		},
	},
	ShortDescription: "loads and follows log files",
}

func runRun(c guinea.Context) error {
	conf := config.Default()
	conf.ServeAddress = c.Options["address"].Str()
	conf.LogFormat = c.Options["log-format"].Str()

	p, err := parser.NewParser(getLogFormat(conf.LogFormat))
	if err != nil {
		return err
	}

	r := core.NewRepository(conf)

	tracker := core.NewTracker(p, r)
	errC := make(chan error)

	// Statistics
	go printStats(tracker)

	// Load the specified files
	for i := 1; i < len(c.Arguments); i++ {
		go func(i int) {
			errC <- tracker.Load(c.Arguments[i])
		}(i)
	}

	for i := 1; i < len(c.Arguments); i++ {
		if err := <-errC; err != nil {
			return err
		}
	}

	// Track the specified file
	go func() {
		errC <- tracker.Follow(c.Arguments[0])
	}()

	go func() {
		errC <- server.Serve(tracker.Repository, conf.ServeAddress)
	}()

	return <-errC
}

// getLogFormat tries to find and return a predefined format with the provided
// name or otherwise returns the provided format unaltered assuming that it is
// a format string.
func getLogFormat(format string) string {
	predefinedFormat, ok := parser.PredefinedFormats[format]
	if ok {
		return predefinedFormat
	}
	return format
}

func printStats(tracker *core.Tracker) {
	lastLines, _ := tracker.GetStats()
	duration := 1 * time.Second
	for range time.Tick(duration) {
		lines, _ := tracker.GetStats()
		linesPerSecond := float64(lines-lastLines) / duration.Seconds()
		log.Debug("data statistics", "totalLines", lines, "linesPerSecond", linesPerSecond)
		lastLines = lines
		logMemoryStats()
	}
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
