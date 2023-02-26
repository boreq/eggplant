//go:build wireinject
// +build wireinject

package wire

import (
	"context"

	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/internal/config"
	"github.com/boreq/eggplant/internal/service"
	"github.com/google/wire"
	bolt "go.etcd.io/bbolt"
)

func BuildTransactableAuthRepositories(tx *bolt.Tx) (*auth.TransactableRepositories, error) {
	wire.Build(
		appSet,
	)

	return nil, nil
}

func BuildTransactableQueryRepositories(tx *bolt.Tx) (*queries.TransactableRepositories, error) {
	wire.Build(
		appSet,
	)

	return nil, nil
}

func BuildAuthForTest(db *bolt.DB) (*auth.Auth, error) {
	wire.Build(
		appSet,
	)

	return nil, nil
}

func BuildAuth(conf *config.Config) (*auth.Auth, error) {
	wire.Build(
		appSet,
		boltSet,
	)

	return nil, nil
}

func BuildService(ctx context.Context, conf *config.Config) (*service.Service, error) {
	wire.Build(
		service.NewService,
		httpSet,
		appSet,
		musicSet,
		boltSet,
	)

	return nil, nil
}
