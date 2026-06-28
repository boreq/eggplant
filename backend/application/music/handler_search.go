package music

import (
	"context"
	"fmt"

	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/eggplant/internal/logging"
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
	repo          LibraryRepository
	remoteLibrary RemoteLibrary
	log           logging.Logger
}

func NewSearchHandler(repo LibraryRepository, remoteLibrary RemoteLibrary) *SearchHandler {
	return &SearchHandler{
		repo:          repo,
		remoteLibrary: remoteLibrary,
		log:           logging.New("music.SearchHandler"),
	}
}

func (h *SearchHandler) Execute(ctx context.Context, accessCtx accessctx.AccessContext, cmd Search) (library.SearchResults, error) {
	if cmd.Query.IsZero() {
		return library.SearchResults{}, errors.New("zero value of query")
	}
	lib, err := h.repo.Get()
	if err != nil {
		return library.SearchResults{}, errors.Wrap(err, "could not get the library")
	}

	localResults, err := lib.Search(accessCtx, cmd.Query.String())
	if err != nil {
		return library.SearchResults{}, errors.Wrap(err, "local search failed")
	}

	if !accessCtx.Can(accessctx.PermissionSeeRemoteLibraryContent) {
		return localResults, nil
	}

	remoteResults, err := h.remoteLibrary.Search(ctx, cmd.Query.String())
	if err != nil {
		h.log.Error("remote search failed", "err", err)
		return localResults, nil
	}

	return library.MergeResults(localResults, remoteResults), nil
}
