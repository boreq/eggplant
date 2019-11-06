package wire

import (
	"path/filepath"

	"github.com/boreq/eggplant/adapters"
	authAdapters "github.com/boreq/eggplant/adapters/auth"
	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/cmd/eggplant/commands/config"
	"github.com/google/wire"
	bolt "go.etcd.io/bbolt"
)

//lint:ignore U1000 because
var appSet = wire.NewSet(
	wire.Struct(new(application.Application), "*"),

	wire.Struct(new(application.Auth), "*"),
	auth.NewRegisterInitialHandler,
	auth.NewLoginHandler,
	auth.NewLogoutHandler,
	auth.NewCheckAccessTokenHandler,
	auth.NewListHandler,
	auth.NewCreateInvitationHandler,
	auth.NewRegisterHandler,
	auth.NewRemoveHandler,

	wire.Struct(new(application.Music), "*"),
	music.NewTrackHandler,
	music.NewThumbnailHandler,
	music.NewBrowseHandler,

	wire.Struct(new(application.Queries), "*"),
	queries.NewStatsHandler,

	wire.Bind(new(queries.UserRepository), new(*authAdapters.UserRepository)),
	wire.Bind(new(auth.UserRepository), new(*authAdapters.UserRepository)),
	authAdapters.NewUserRepository,

	wire.Bind(new(authAdapters.PasswordHasher), new(*authAdapters.BcryptPasswordHasher)),
	authAdapters.NewBcryptPasswordHasher,

	wire.Bind(new(authAdapters.AccessTokenGenerator), new(*authAdapters.CryptoAccessTokenGenerator)),
	authAdapters.NewCryptoAccessTokenGenerator,

	newBolt,
)

func newBolt(conf *config.Config) (*bolt.DB, error) {
	path := filepath.Join(conf.DataDirectory, "eggplant.database")
	return adapters.NewBolt(path)
}
