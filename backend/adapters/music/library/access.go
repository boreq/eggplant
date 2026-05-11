package library

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/errors"
)

type DelimiterAccessLoader struct{}

func NewDelimiterAccessLoader() *DelimiterAccessLoader {
	return &DelimiterAccessLoader{}
}

func (l *DelimiterAccessLoader) Load(file string) (music.Access, error) {
	f, err := os.Open(file)
	if err != nil {
		return music.Access{}, errors.Wrap(err, "could not open the file")
	}
	defer f.Close()

	acc := music.Access{
		Public: false,
	}

	emptyFile := true
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		emptyFile = false
		key, value, err := l.loadLine(line)
		if err != nil {
			return music.Access{}, errors.Wrap(err, "could not parse a line")
		}
		switch key {
		case "public":
			acc.Public = value
		default:
			return music.Access{}, fmt.Errorf("unrecognized key '%s'", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return music.Access{}, errors.Wrap(err, "scanner error")
	}

	if emptyFile {
		return music.Access{}, fmt.Errorf("access file is empty: '%s'", file)
	}

	return acc, nil
}

func (l *DelimiterAccessLoader) loadLine(line string) (string, bool, error) {
	line = strings.TrimSpace(line)
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
