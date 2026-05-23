package titleparser

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type NumberAndTitle struct {
	number *domain.TrackNumber
	title  domain.TrackTitle
}

func NewNumberAndTitle(number *domain.TrackNumber, title domain.TrackTitle) NumberAndTitle {
	return NumberAndTitle{
		number: number,
		title:  title,
	}
}

func (p NumberAndTitle) Title() domain.TrackTitle {
	return p.title
}

func (p NumberAndTitle) Number() *domain.TrackNumber {
	return p.number
}

type state int

const (
	stateStart state = iota
	stateNumber
	stateSeparator
	stateTitle
)

func Parse(s string) (NumberAndTitle, error) {
	numberAndTitle, err := tryParse(s)
	if err != nil {
		title, err := domain.NewTrackTitle(s)
		if err != nil {
			return NumberAndTitle{}, errors.Wrap(err, "error creating a fallback track title")
		}
		return NewNumberAndTitle(nil, title), nil
	}
	return numberAndTitle, nil
}

func tryParse(s string) (NumberAndTitle, error) {
	st := stateStart
	var numberBuf, titleBuf strings.Builder

	for _, r := range s {
		switch st {
		case stateStart:
			switch {
			case unicode.IsSpace(r):
			case unicode.IsDigit(r):
				numberBuf.WriteRune(r)
				st = stateNumber
			default:
				titleBuf.WriteRune(r)
				st = stateTitle
			}
		case stateNumber:
			if unicode.IsDigit(r) {
				numberBuf.WriteRune(r)
			} else if isSeparator(r) {
				st = stateSeparator
			} else {
				return NumberAndTitle{}, errors.New("digits not followed by separator")
			}
		case stateSeparator:
			if !isSeparator(r) {
				titleBuf.WriteRune(r)
				st = stateTitle
			}
		case stateTitle:
			titleBuf.WriteRune(r)
		}
	}

	return tryCreatingNumberAndTitle(numberBuf.String(), titleBuf.String())
}

func tryCreatingNumberAndTitle(numStr, titleStr string) (NumberAndTitle, error) {
	var number *domain.TrackNumber
	if numStr != "" {
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return NumberAndTitle{}, errors.Wrap(err, "could not parse number")
		}
		num, err := domain.NewTrackNumber(n)
		if err != nil {
			return NumberAndTitle{}, errors.Wrap(err, "could not create track number")
		}
		number = &num
	}

	title, err := domain.NewTrackTitle(titleStr)
	if err != nil {
		return NumberAndTitle{}, errors.Wrap(err, "could not create track title")
	}

	return NewNumberAndTitle(number, title), nil
}

func isSeparator(r rune) bool {
	return r == ' ' || r == '\t' || r == '.' || r == '-'
}
