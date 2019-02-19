package core

import (
	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser"
	"github.com/hpcloud/tail"
)

func NewTracker(p *parser.Parser) *Tracker {
	return &Tracker{
		Repository: NewRepository(),
		parser:     p,
		log:        logging.New("core/tracker"),
	}
}

type Tracker struct {
	Repository *Repository
	parser     *parser.Parser
	Lines      int
	Errors     int
	log        logging.Logger
}

func (t *Tracker) Follow(filepath string) error {
	t.log.Debug("following file", "filepath", filepath)
	ta, err := tail.TailFile(filepath, tail.Config{Follow: true})
	if err != nil {
		return err
	}
	return t.processFile(ta)
}

func (t *Tracker) Load(filepath string) error {
	t.log.Debug("loading file", "filepath", filepath)
	ta, err := tail.TailFile(filepath, tail.Config{MustExist: true})
	if err != nil {
		return err
	}
	return t.processFile(ta)
}

func (t *Tracker) processFile(ta *tail.Tail) error {
	for line := range ta.Lines {
		t.Lines++
		err := t.processLine(line.Text)
		if err != nil {
			t.Errors++
			t.log.Error("error processing a line", "err", err, "line", line.Text)
		}
	}
	return nil
}

func (t *Tracker) processLine(line string) error {
	entry, err := t.parser.Parse(line)
	if err != nil {
		return err
	}
	return t.Repository.Insert(entry)
}
