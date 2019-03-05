// Package parser handles log parsing.
package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/boreq/plum/logging"
	"github.com/boreq/plum/parser/format"
	"github.com/pkg/errors"
)

var log = logging.New("parser")

type Entry struct {
	// Client address.
	RemoteAddress string

	// Local time.
	Time time.Time

	// HTTP request line: method.
	HttpRequestMethod string

	// HTTP request line: URI.
	HttpRequestURI string

	// HTTP request line: version.
	HttpRequestVersion string

	// Response status.
	Status string

	// Body bytes sent.
	BodyBytesSent int

	// HTTP referer.
	Referer string

	// User agent.
	UserAgent string
}

var PredefinedFormats = map[string]string{
	// https://nginx.org/en/docs/http/ngx_http_log_module.html#log_format
	"combined": "${remote_addr} - ${remote_user} [${time_local}] \"${request}\" ${status} ${body_bytes_sent} \"${http_referer}\" \"${http_user_agent}\"",
}

// parseRemoteAddr places the provided string in the RemoteAddress field of the
// entry without modifying it. This should be an IP address of the client.
//
// Input example: "11.22.33.44"
func parseRemoteAddr(entry *Entry, s string) error {
	entry.RemoteAddress = s
	return nil
}

// parseTimeLocal treats the provided string as a time saved in the Common Log
// Format and stores it in the Time field of the entry.
//
// Input example: "03/Mar/2019:00:01:36 +0100"
func parseTimeLocal(entry *Entry, s string) error {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", s)
	if err != nil {
		return err
	}
	entry.Time = t
	return nil
}

// parseRequest treats the provided string as an HTTP request line and
// populates all the entry fields having the names starting with HttpRequest*.
// If the request line is empty no error is returned as this can be the case if
// this entry in the log file was an invalid request.
//
// https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html#sec5.1
//
// Input example: "GET / HTTP/2.0"
func parseRequest(entry *Entry, s string) error {
	spl := strings.SplitN(s, " ", 3)
	if len(spl) != 3 {
		return nil
	}
	entry.HttpRequestMethod = spl[0]
	entry.HttpRequestURI = spl[1]
	entry.HttpRequestVersion = spl[2]
	return nil
}

// parseStatus places the provided string in the Status field of the entry
// without modifying it.
//
// Input example: "200"
func parseStatus(entry *Entry, s string) error {
	entry.Status = s
	return nil
}

// parseBodyBytesSent treats the provided string as a number and places it in
// the BodyBytesSent field of the entry. The number should represent the number
// of bytes transferred to the client.
//
// Input example: "123"
func parseBodyBytesSent(entry *Entry, s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	entry.BodyBytesSent = n
	return nil
}

// parseHttpReferer places the provided string in the Referer field of the
// entry without modifying it. This should be an HTTP referer. If a referer is
// missing this log field is often filled with a dash ("-") instead. This is
// transferred without modification as well.
//
// Input example: "https://www.reddit.com/r/programming/comments/123456/some-title/"
func parseHttpReferer(entry *Entry, s string) error {
	entry.Referer = s
	return nil
}

// parseUserAgent places the provided string in the UserAgent field of the
// entry without modifying it. This should be an HTTP User-Agent.
//
// Input example: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.119 Safari/537.36"
func parseUserAgent(entry *Entry, s string) error {
	entry.UserAgent = s
	return nil
}

// doNothing is a placeholder which... does nothing. It can be used to consume
// a part of the log without using it. This is useful if parsing a specific
// value is not supported.
func doNothing(entry *Entry, s string) error {
	return nil
}

type parseFn func(entry *Entry, s string) error

var parseFunctions = map[string]parseFn{
	"remote_addr":     parseRemoteAddr,
	"time_local":      parseTimeLocal,
	"request":         parseRequest,
	"status":          parseStatus,
	"body_bytes_sent": parseBodyBytesSent,
	"http_referer":    parseHttpReferer,
	"http_user_agent": parseUserAgent,

	"remote_user": doNothing,

	"do_nothing": doNothing,
}

func validate(items []format.Item) error {
	// There should never be two consecutive items with the same ItemType
	// in the format, otherwise it wouldn't be possible to differentiate
	// between two subsequent format items
	for i := 0; i < len(items)-1; i++ {
		if items[i].Type == items[i+1].Type {
			errors.New("format elements must be delimited with a static string")
		}
	}

	// All parse functions defined in the format should exist
	for _, item := range items {
		if item.Type == format.ItemElement {
			_, ok := parseFunctions[item.Value]
			if !ok {
				return fmt.Errorf("unknown parse function %s", item.Value)
			}
		}
	}

	return nil
}

type Parser struct {
	formatItems []format.Item
}

// NewParser creates a new parser using the specified format.
func NewParser(frmt string) (*Parser, error) {
	formatItems, err := format.Lex(frmt)
	if err != nil {
		return nil, errors.Wrap(err, "format lexing failed")
	}

	err = validate(formatItems)
	if err != nil {
		return nil, errors.Wrap(err, "format validation failed")
	}

	rv := &Parser{
		formatItems: formatItems,
	}
	return rv, nil
}

// Parse parses a line of the log file.
func (p *Parser) Parse(line string) (*Entry, error) {
	entry := &Entry{}
	for i := range p.formatItems {
		// Line is now empty but we still have formatItems that want to
		// consume some values
		if line == "" {
			return nil, errors.New("ran out of input")
		}

		item := p.formatItems[i]
		switch item.Type {

		// If this is a text delimiter simply consume and discard it
		case format.ItemText:
			if !strings.HasPrefix(line, item.Value) {
				return nil, fmt.Errorf("expected delimiter \"%s\" at the beginning of input \"%s\"", item.Value, line)
			}
			line = strings.TrimPrefix(line, item.Value)

		// If this is an item that needs to be parsed try to find a
		// related parse function and use it to parse this element
		case format.ItemElement:
			parseFunction, ok := parseFunctions[item.Value]
			if !ok {
				return nil, fmt.Errorf("unknown parse function ${%s}", item.Value)
			}
			if i+1 < len(p.formatItems) {
				nextItem := p.formatItems[i+1]
				if nextItem.Type != format.ItemText {
					return nil, errors.New("expected a delimiter in the format")
				}
				i := strings.Index(line, nextItem.Value)
				if i < 0 {
					return nil, errors.New("expected a delimiter in the input")
				}
				err := parseFunction(entry, line[:i])
				if err != nil {
					return nil, errors.Wrapf(err, "error in ${%s}", item.Value)
				}
				line = line[i:]
			} else {
				err := parseFunction(entry, line)
				if err != nil {
					return nil, errors.Wrapf(err, "error in ${%s}", item.Value)
				}
				line = ""
			}
		}
	}
	return entry, nil
}
