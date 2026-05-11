package music

import (
	"errors"
	"fmt"
)

const maxQueryLength = 100

type Query struct {
	s string
}

func NewQuery(s string) (Query, error) {
	if s == "" {
		return Query{}, errors.New("query can not be empty")
	}

	if len(s) > maxQueryLength {
		return Query{}, fmt.Errorf("query can not be longer than %d", maxQueryLength)
	}

	return Query{
		s: s,
	}, nil
}

func MustNewQuery(s string) Query {
	v, err := NewQuery(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (q Query) IsZero() bool {
	return q == Query{}
}

func (q Query) String() string {
	return q.s
}

type Search struct {
	Query      Query
	PublicOnly bool
}

type SearchHandler struct {
	library Library
}

func NewSearchHandler(library Library) *SearchHandler {
	return &SearchHandler{
		library: library,
	}
}

func (h *SearchHandler) Execute(cmd Search) (SearchResult, error) {
	if cmd.Query.IsZero() {
		return SearchResult{}, errors.New("zero value of query")
	}

	return h.library.Search(cmd.Query.String(), cmd.PublicOnly)
}
