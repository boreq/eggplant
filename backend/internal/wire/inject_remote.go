package wire

import (
	"github.com/boreq/eggplant/adapters"
	"github.com/boreq/eggplant/adapters/pubsub"
	remoteAdapters "github.com/boreq/eggplant/adapters/remote"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/remote"
	outboxEntrypoint "github.com/boreq/eggplant/entrypoints/outbox"
	"github.com/google/wire"
	bolt "go.etcd.io/bbolt"
)

//lint:ignore U1000 because
var remoteSet = wire.NewSet(
	wire.Struct(new(remote.Remote), "*"),
	remote.NewAddRemoteHandler,
	remote.NewSetRemotePairingTokenHandler,
	remote.NewSendLocalAuthTokenHandler,
	remote.NewSetRemoteAuthTokenHandler,
	remote.NewCheckLocalAuthTokenHandler,
	remote.NewListRemotesHandler,
	remote.NewListRemoteLibrariesHandler,
	remote.NewCheckRemotesHandler,
	remote.NewCheckRemoteHandler,

	remoteAdapters.NewRemoteTransactionProvider,
	wire.Bind(new(remote.TransactionProvider), new(*remoteAdapters.RemoteTransactionProvider)),

	remoteAdapters.NewRemoteClient,
	wire.Bind(new(remote.RemoteClient), new(*remoteAdapters.RemoteClient)),

	remoteAdapters.NewRemoteLibrary,
	wire.Bind(new(music.RemoteLibrary), new(*remoteAdapters.RemoteLibrary)),

	newRemoteRepositoriesProvider,
	wire.Bind(new(remoteAdapters.RemoteRepositoriesProvider), new(*remoteRepositoriesProvider)),

	pubsub.NewPubSub,

	outboxEntrypoint.NewListener,
	wire.Bind(new(outboxEntrypoint.Subscriber), new(*pubsub.PubSub)),
	wire.Bind(new(outboxEntrypoint.SendLocalAuthTokenHandler), new(*remote.SendLocalAuthTokenHandler)),
	wire.Bind(new(outboxEntrypoint.CheckRemoteHandler), new(*remote.CheckRemoteHandler)),

	wire.Struct(new(remote.TransactableRepositories), "*"),
	wire.Bind(new(remote.RemoteInstanceRepository), new(*remoteAdapters.RemoteInstanceRepository)),
	remoteAdapters.NewRemoteInstanceRepository,
	wire.Bind(new(remote.OutboxRepository), new(*adapters.OutboxRepository)),
	adapters.NewOutboxRepository,
)

type remoteRepositoriesProvider struct {
}

func newRemoteRepositoriesProvider() *remoteRepositoriesProvider {
	return &remoteRepositoriesProvider{}
}

func (p *remoteRepositoriesProvider) Provide(tx *bolt.Tx) (*remote.TransactableRepositories, error) {
	return BuildTransactableRemoteRepositories(tx)
}
