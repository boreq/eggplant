package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser/format"
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

func parseRemoteAddr(entry *Entry, s string) error {
	entry.RemoteAddress = s
	return nil
}

func parseTimeLocal(entry *Entry, s string) error {
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", s)
	if err != nil {
		return err
	}
	entry.Time = t
	return nil
}

func parseRequest(entry *Entry, s string) error {
	spl := strings.SplitN(s, " ", 3)
	if len(spl) != 3 {
		log.Warn("request line is invalid, error is not being returned as technically it is possible that a client made an invalid request which was logged - if you see a lot of those warnings then your parsing format is most likely incorrect", "invalid_request_line", s)
		return nil
	}
	entry.HttpRequestMethod = spl[0]
	entry.HttpRequestURI = spl[1]
	entry.HttpRequestVersion = spl[2]
	return nil
}

func parseStatus(entry *Entry, s string) error {
	entry.Status = s
	return nil
}

func parseBodyBytesSent(entry *Entry, s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	entry.BodyBytesSent = n
	return nil
}

func parseHttpReferer(entry *Entry, s string) error {
	entry.Referer = s
	return nil
}

func parseUserAgent(entry *Entry, s string) error {
	entry.UserAgent = s
	return nil
}

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
}

func validate(items []format.Item) error {
	for i := 0; i < len(items)-1; i++ {
		if items[i].Type == items[i+1].Type {
			errors.New("format elements must be delimited with at least one character")
		}
	}
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
		errors.Wrap(err, "invalid format")
	}

	err = validate(formatItems)
	if err != nil {
		errors.Wrap(err, "invalid format")
	}

	rv := &Parser{
		formatItems: formatItems,
	}
	return rv, nil
}

func (p *Parser) Parse(line string) (*Entry, error) {
	entry := &Entry{}
	for i := range p.formatItems {
		if line == "" {
			return nil, errors.New("ran out of input")
		}
		item := p.formatItems[i]
		switch item.Type {
		case format.ItemText:
			if !strings.HasPrefix(line, item.Value) {
				return nil, fmt.Errorf("expected prefix \"%s\" in input \"%s\"", item.Value, line)
			}
			line = strings.TrimPrefix(line, item.Value)
		case format.ItemElement:
			parseFunction, ok := parseFunctions[item.Value]
			if !ok {
				return nil, fmt.Errorf("unknown parse function %s", item.Value)
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
