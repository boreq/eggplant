package core

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"strings"
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
	return t.processTail(ta)
}

func (t *Tracker) Load(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.processFile(f)
}

func (t *Tracker) processFile(f *os.File) error {
	r, err := t.getReader(f)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		t.addLine()
		err := t.processLine(scanner.Text())
		if err != nil {
			t.addError()
			t.log.Error("error processing a line", "err", err, "line", scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (t *Tracker) getReader(f *os.File) (io.Reader, error) {
	if strings.HasSuffix(f.Name(), ".gz") {
		r, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	return f, nil
}

func (t *Tracker) processTail(ta *tail.Tail) error {
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
