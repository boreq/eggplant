package library

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type DelimiterAccessLoader struct{}

func NewDelimiterAccessLoader() *DelimiterAccessLoader {
	return &DelimiterAccessLoader{}
}

func (l *DelimiterAccessLoader) Load(file string) (library.Visibility, error) {
	f, err := os.Open(file)
	if err != nil {
		return library.Visibility{}, errors.Wrap(err, "could not open the file")
	}
	defer f.Close()

	var public *bool
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, err := l.loadLine(line)
		if err != nil {
			return library.Visibility{}, errors.Wrap(err, "could not parse a line")
		}
		switch key {
		case "public":
			if public != nil {
				return library.Visibility{}, fmt.Errorf("duplicate key '%s'", key)
			}
			public = &value
		default:
			return library.Visibility{}, fmt.Errorf("unrecognized key '%s'", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return library.Visibility{}, errors.Wrap(err, "scanner error")
	}

	if public == nil {
		return library.Visibility{}, errors.New("access file is empty")
	}

	return library.NewVisibility(*public), nil
}

func (l *DelimiterAccessLoader) loadLine(line string) (string, bool, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", false, fmt.Errorf("malformed line '%s'", line)
	}
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])
	var value bool
	switch parts[1] {
	case "yes":
		value = true
	case "no":
		value = false
	default:
		return "", false, fmt.Errorf("value '%s' is not 'yes' or 'no'", parts[1])
	}
	return parts[0], value, nil
}
