package music

import (
	"fmt"

	library2 "github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
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
	Query Query
}

type SearchHandler struct {
	repo LibraryRepository
}

func NewSearchHandler(repo LibraryRepository) *SearchHandler {
	return &SearchHandler{repo: repo}
}

func (h *SearchHandler) Execute(accessCtx library2.AccessContext, cmd Search) (library2.SearchResults, error) {
	if cmd.Query.IsZero() {
		return library2.SearchResults{}, errors.New("zero value of query")
	}
	lib, err := h.repo.Get()
	if err != nil {
		return library2.SearchResults{}, errors.Wrap(err, "could not get the library")
	}
	return lib.Search(accessCtx, cmd.Query.String())
}
