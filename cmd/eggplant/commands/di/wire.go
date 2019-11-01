//+build wireinject

package di

import (
	"github.com/boreq/eggplant/config"
	"github.com/boreq/eggplant/library"
	"github.com/boreq/eggplant/store"
	"github.com/google/wire"
)

func BuildService(lib *library.Library, trackStore *store.TrackStore, thumbnailStore *store.Store, conf *config.Config) (*Service, error) {
	wire.Build(
		NewService,
		httpSet,
		appSet,
	)

	return nil, nil
}
