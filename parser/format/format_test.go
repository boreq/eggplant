package format

import (
	"testing"
)

type testCase struct {
	input  string
	output []Item
}

var testCases = []testCase{
	{
		input:  "",
		output: []Item{},
	},
	{
		input: "${element}",
		output: []Item{
			{ItemElement, "element"},
		},
	},
	{
		input: "text",
		output: []Item{
			{ItemText, "text"},
		},
	},
	{
		input: "text${element}",
		output: []Item{
			{ItemText, "text"},
			{ItemElement, "element"},
		},
	},
	{
		input: "${element}text",
		output: []Item{
			{ItemElement, "element"},
			{ItemText, "text"},
		},
	},
	{
		input: "${element}text${element}",
		output: []Item{
			{ItemElement, "element"},
			{ItemText, "text"},
			{ItemElement, "element"},
		},
	},
	{
		input: "text${element}text",
		output: []Item{
			{ItemText, "text"},
			{ItemElement, "element"},
			{ItemText, "text"},
		},
	},

	// escape related test cases
	{
		input: "\\${element\\}${element}text",
		output: []Item{
			{ItemText, "${element}"},
			{ItemElement, "element"},
			{ItemText, "text"},
		},
	},
	{
		input: "\\${element}${element}text",
		output: []Item{
			{ItemText, "${element}"},
			{ItemElement, "element"},
			{ItemText, "text"},
		},
	},
	{
		input: "text\\",
		output: []Item{
			{ItemError, "text\\"},
		},
	},
	{
		input: "${element\\}}",
		output: []Item{
			{ItemElement, "element}"},
		},
	},
}

func TestLex(t *testing.T) {
	for _, test := range testCases {
		ch := lex(test.input)

		for _, expected := range test.output {
			actual, ok := <-ch
			if !ok {
				t.Fatalf("input ended, expected: %#v", expected)
			}

			if actual != expected {
				t.Fatalf("actual: %#v, expected: %#v", actual, expected)
			}

		}

		unexpected, ok := <-ch
		if ok {
			t.Fatalf("actual items should have ended, but got: %#v", unexpected)
		}
	}
}
