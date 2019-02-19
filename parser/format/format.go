package format

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type ItemType int

const (
	ItemError ItemType = iota

	ItemElement
	ItemText
)

type Item struct {
	Type  ItemType
	Value string
}

const escape = '\\'
const startMarker = "${"
const endMarker = "}"

type stateFn func(*lexer) stateFn

var lexText stateFn

var lexElement stateFn

func init() {
	lexText = lexUntil(startMarker, ItemText, lexStartMarker, true)
	lexElement = lexUntil(endMarker, ItemElement, lexEndMarker, false)
}

func lexUntil(until string, emitItemType ItemType, nextState stateFn, eofAllowed bool) stateFn {
	return func(l *lexer) stateFn {
		for {
			if strings.HasPrefix(l.followingInput(), until) {
				if l.pos > l.start {
					l.emit(emitItemType)
				}
				return nextState
			}
			r, err := l.next()
			if err != nil {
				if eofAllowed {
					break
				} else {
					l.emit(ItemError)
				}
			}
			if r == escape {
				next := lexUntil(until, emitItemType, nextState, eofAllowed)
				return lexEscape(next)
			}
		}
		if l.pos > l.start {
			l.emit(emitItemType)
		}
		return nil
	}
}

func lexEscape(nextState stateFn) stateFn {
	return func(l *lexer) stateFn {
		// Read the next rune to consume it
		r, err := l.next()
		if err != nil {
			l.emit(ItemError)
			return nil
		}

		// Remove the "escape" rune from the input and move the
		// position pointer back to compensate for this
		width := len(string(escape))
		start := l.pos - 1 - len(string(r))
		l.input = l.input[:start] + l.input[start+width:]
		l.pos -= width

		return nextState
	}
}

func lexStartMarker(l *lexer) stateFn {
	l.pos += len(startMarker)
	l.start = l.pos
	return lexElement
}

func lexEndMarker(l *lexer) stateFn {
	l.pos += len(endMarker)
	l.start = l.pos
	return lexText
}

type lexer struct {
	input string
	pos   int
	start int
	items chan Item
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) followingInput() string {
	return l.input[l.pos:]
}

func (l *lexer) next() (rune, error) {
	if l.pos >= len(l.input) {
		return 0, io.EOF
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	return r, nil

}

func (l *lexer) emit(t ItemType) {
	l.items <- Item{Type: t, Value: l.input[l.start:l.pos]}
	l.start = l.pos
}

func lex(format string) <-chan Item {
	l := &lexer{
		input: format,
		items: make(chan Item),
	}
	go l.run()
	return l.items
}

func Lex(format string) ([]Item, error) {
	ch := lex(format)
	var items []Item
	for item := range ch {
		switch item.Type {
		case ItemError:
			return nil, fmt.Errorf("unexpected: %s", item.Value)
		default:
			items = append(items, item)
		}
	}
	return items, nil
}
