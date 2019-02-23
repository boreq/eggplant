package core

import (
	"sync"

	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser"
	"github.com/hpcloud/tail"
)

func NewTracker(p *parser.Parser, r *Repository) *Tracker {
	return &Tracker{
		Repository: r,
		parser:     p,
		log:        logging.New("core/tracker"),
	}
}

type Tracker struct {
	Repository *Repository
	parser     *parser.Parser
	log        logging.Logger

	lines      int
	errors     int
	statsMutex sync.Mutex
}

func (t *Tracker) Follow(filepath string) error {
	ta, err := tail.TailFile(filepath, tail.Config{Follow: true})
	if err != nil {
		return err
	}
	return t.processFile(ta)
}

func (t *Tracker) Load(filepath string) error {
	ta, err := tail.TailFile(filepath, tail.Config{MustExist: true})
	if err != nil {
		return err
	}
	return t.processFile(ta)
}

func (t *Tracker) processFile(ta *tail.Tail) error {
	for line := range ta.Lines {
		t.addLine()
		err := t.processLine(line.Text)
		if err != nil {
			t.addError()
			t.log.Error("error processing a line", "err", err, "line", line.Text)
		}
	}
	return nil
}

func (t *Tracker) addLine() {
	t.statsMutex.Lock()
	defer t.statsMutex.Unlock()
	t.lines++
}

func (t *Tracker) addError() {
	t.statsMutex.Lock()
	defer t.statsMutex.Unlock()
	t.errors++
}

func (t *Tracker) GetStats() (lines int, errors int) {
	t.statsMutex.Lock()
	defer t.statsMutex.Unlock()
	lines = t.lines
	errors = t.errors
	return
}

func (t *Tracker) processLine(line string) error {
	entry, err := t.parser.Parse(line)
	if err != nil {
		return err
	}
	return t.Repository.Insert(entry)
}
