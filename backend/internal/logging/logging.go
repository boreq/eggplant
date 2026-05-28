// Package logging provides a layer of abstraction around external loggers.
package logging

import (
	"github.com/inconshreveable/log15"
)

type Logger = log15.Logger

type Level = log15.Lvl

var maxLevel *Level

func init() {
	level := log15.LvlDebug
	maxLevel = &level
}

func New(name string) Logger {
	log := log15.New("source", name)
	log.SetHandler(log15.FilterHandler(func(r *log15.Record) (pass bool) {
		return r.Lvl <= *maxLevel
	}, log15.StdoutHandler))
	return log
}

func SetLoggingLevel(level Level) {
	maxLevel = &level
}

func LevelFromString(s string) (Level, error) {
	return log15.LvlFromString(s)
}
