package parser

import (
	"testing"
)

func TestCombined(t *testing.T) {
	var lines = []string{
		"12.123.1.123 - - [03/Feb/2019:00:00:22 +0100] \"GET /static/css/main.css HTTP/2.0\" 200 10267 \"https://0x46.net/thoughts/2019/02/01/dotfile-madness/\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36\"",
	}

	frmt := PredefinedFormats["combined"]
	p, err := NewParser(frmt)
	if err != nil {
		t.Fatal(err)
	}

	for _, line := range lines {
		entry, err := p.Parse(line)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("entry %#v", entry)
	}
}
