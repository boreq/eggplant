package wire

import (
	authAdapters "github.com/boreq/eggplant/adapters/auth"
	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
	"github.com/boreq/eggplant/entrypoints/filesystem"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/eggplant/internal/version"
	"github.com/google/wire"
	bolt "go.etcd.io/bbolt"
)

func newMusicHandlerLogger() logging.Logger {
	return logging.New("music.handler")
}

//lint:ignore U1000 because
var appSet = wire.NewSet(
	wire.Struct(new(application.Application), "*"),

	wire.Struct(new(auth.Auth), "*"),
	auth.NewRegisterInitialHandler,
	auth.NewLoginHandler,
	auth.NewLogoutHandler,
	auth.NewCheckAccessTokenHandler,
	auth.NewListHandler,
	auth.NewCreateInvitationHandler,
	auth.NewRegisterHandler,
	auth.NewRemoveHandler,
	auth.NewSetPasswordHandler,
	auth.NewGetCurrentUserHandler,

	wire.Struct(new(music.Music), "*"),
	newMusicHandlerLogger,
	music.NewStartStreamingHandler,
	music.NewStreamPlaylistHandler,
	music.NewStreamInitHandler,
	music.NewStreamFragmentHandler,
	music.NewKeepAliveStreamHandler,
	music.NewThumbnailHandler,
	music.NewGetRootAlbumHandler,
	music.NewGetAlbumHandler,
	music.NewRemoteGetAlbumHandler,
	music.NewRemoteGetThumbnailHandler,
	music.NewRemoteGetTrackDurationHandler,
	music.NewRemoteStartStreamingHandler,
	music.NewRemoteStreamPlaylistHandler,
	music.NewRemoteStreamInitHandler,
	music.NewRemoteStreamFragmentHandler,
	music.NewRemoteKeepAliveStreamHandler,
	music.NewGetTrackDurationHandler,
	music.NewSearchHandler,
	music.NewLoadLibraryHandler,
	music.NewLoggingStartStreamingHandler,
	music.NewLoggingStreamPlaylistHandler,
	music.NewLoggingStreamInitHandler,
	music.NewLoggingStreamFragmentHandler,
	music.NewLoggingKeepAliveStreamHandler,
	music.NewLoggingThumbnailHandler,
	music.NewLoggingGetRootAlbumHandler,
	music.NewLoggingGetAlbumHandler,
	music.NewLoggingRemoteGetAlbumHandler,
	music.NewLoggingRemoteGetThumbnailHandler,
	music.NewLoggingRemoteGetTrackDurationHandler,
	music.NewLoggingRemoteStartStreamingHandler,
	music.NewLoggingRemoteStreamPlaylistHandler,
	music.NewLoggingRemoteStreamInitHandler,
	music.NewLoggingRemoteStreamFragmentHandler,
	music.NewLoggingRemoteKeepAliveStreamHandler,
	music.NewLoggingGetTrackDurationHandler,
	music.NewLoggingSearchHandler,
	music.NewLoggingLoadLibraryHandler,
	wire.Bind(new(filesystem.LoadLibraryHandler), new(*music.LoggingLoadLibraryHandler)),

	wire.Struct(new(application.Queries), "*"),
	queries.NewStatsHandler,
	newVersionHandler,

	authAdapters.NewAuthTransactionProvider,
	wire.Bind(new(auth.TransactionProvider), new(*authAdapters.AuthTransactionProvider)),

	authAdapters.NewQueryTransactionProvider,
	wire.Bind(new(queries.TransactionProvider), new(*authAdapters.QueryTransactionProvider)),

	wire.Struct(new(auth.TransactableRepositories), "*"),
	wire.Struct(new(queries.TransactableRepositories), "*"),

	newQueryRepositoriesProvider,
	wire.Bind(new(authAdapters.QueryRepositoriesProvider), new(*queryRepositoriesProvider)),

	newAuthRepositoriesProvider,
	wire.Bind(new(authAdapters.AuthRepositoriesProvider), new(*authRepositoriesProvider)),

	wire.Bind(new(queries.UserRepository), new(*authAdapters.UserRepository)),
	wire.Bind(new(auth.UserRepository), new(*authAdapters.UserRepository)),
	authAdapters.NewUserRepository,

	wire.Bind(new(auth.InvitationRepository), new(*authAdapters.InvitationRepository)),
	authAdapters.NewInvitationRepository,

	wire.Bind(new(auth.PasswordHasher), new(*authAdapters.BcryptPasswordHasher)),
	authAdapters.NewBcryptPasswordHasher,

	authAdapters.NewLastSeenUpdater,
	wire.Bind(new(auth.LastSeenUpdater), new(*authAdapters.LastSeenUpdater)),
	auth.NewPersistLastSeenHandler,
)

type authRepositoriesProvider struct {
}

func newAuthRepositoriesProvider() *authRepositoriesProvider {
	return &authRepositoriesProvider{}
}

func (p *authRepositoriesProvider) Provide(tx *bolt.Tx) (*auth.TransactableRepositories, error) {
	return BuildTransactableAuthRepositories(tx)
}

func newVersionHandler() *queries.VersionHandler {
	return queries.NewVersionHandler(version.Current)
}

type queryRepositoriesProvider struct {
}

func newQueryRepositoriesProvider() *queryRepositoriesProvider {
	return &queryRepositoriesProvider{}
}

func (p *queryRepositoriesProvider) Provide(tx *bolt.Tx) (*queries.TransactableRepositories, error) {
	return BuildTransactableQueryRepositories(tx)
}
