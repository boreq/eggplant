//+build wireinject

package wire

import (
	"github.com/boreq/eggplant/cmd/eggplant/commands/config"
	"github.com/boreq/eggplant/cmd/eggplant/commands/service"
	"github.com/google/wire"
)

func BuildService(conf *config.Config) (*service.Service, error) {
	wire.Build(
		service.NewService,
		httpSet,
		appSet,
		musicSet,
	)

	return nil, nil
}
