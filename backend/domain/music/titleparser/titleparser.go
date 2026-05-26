package titleparser

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/errors"
)

type NumberAndTitle struct {
	number *music.TrackNumber
	title  music.TrackTitle
}

func NewNumberAndTitle(number *music.TrackNumber, title music.TrackTitle) NumberAndTitle {
	return NumberAndTitle{
		number: number,
		title:  title,
	}
}

func (p NumberAndTitle) Title() music.TrackTitle {
	return p.title
}

func (p NumberAndTitle) Number() *music.TrackNumber {
	return p.number
}

func Parse(t music.TrackTitle) (NumberAndTitle, error) {
	numberAndTitle, err := tryParse(t.String())
	if err != nil {
		return NewNumberAndTitle(nil, t), nil
	}
	return numberAndTitle, nil
}

func tryParse(s string) (NumberAndTitle, error) {
	acc := &parseAccumulator{}
	st := parseState(parseStateStart)

	for _, r := range s {
		next, err := st(acc, r)
		if err != nil {
			return NumberAndTitle{}, err
		}
		st = next
	}

	return tryCreatingNumberAndTitle(acc.number.String(), acc.title.String())
}

func tryCreatingNumberAndTitle(numStr, titleStr string) (NumberAndTitle, error) {
	var number *music.TrackNumber
	if numStr != "" {
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return NumberAndTitle{}, errors.Wrap(err, "could not parse number")
		}
		num, err := music.NewTrackNumber(n)
		if err != nil {
			return NumberAndTitle{}, errors.Wrap(err, "could not create track number")
		}
		number = &num
	}

	title, err := music.NewTrackTitle(titleStr)
	if err != nil {
		return NumberAndTitle{}, errors.Wrap(err, "could not create track title")
	}

	return NewNumberAndTitle(number, title), nil
}

type parseAccumulator struct {
	number strings.Builder
	title  strings.Builder
}

type parseState func(acc *parseAccumulator, r rune) (parseState, error)

func parseStateStart(acc *parseAccumulator, r rune) (parseState, error) {
	switch {
	case unicode.IsSpace(r):
		return parseStateStart, nil
	case unicode.IsDigit(r):
		acc.number.WriteRune(r)
		return parseStateNumber, nil
	default:
		acc.title.WriteRune(r)
		return parseStateTitle, nil
	}
}

func parseStateNumber(acc *parseAccumulator, r rune) (parseState, error) {
	switch {
	case unicode.IsDigit(r):
		acc.number.WriteRune(r)
		return parseStateNumber, nil
	case isSeparator(r):
		return parseStateSeparator, nil
	default:
		return nil, errors.New("digits not followed by separator")
	}
}

func parseStateSeparator(acc *parseAccumulator, r rune) (parseState, error) {
	if isSeparator(r) {
		return parseStateSeparator, nil
	}
	acc.title.WriteRune(r)
	return parseStateTitle, nil
}

func parseStateTitle(acc *parseAccumulator, r rune) (parseState, error) {
	acc.title.WriteRune(r)
	return parseStateTitle, nil
}

func isSeparator(r rune) bool {
	return r == ' ' || r == '\t' || r == '.' || r == '-'
}
