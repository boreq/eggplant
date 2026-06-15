package wire

import (
	"net/http"

	"github.com/boreq/eggplant/adapters/remotes"
	httpPort "github.com/boreq/eggplant/entrypoints/http"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var httpSet = wire.NewSet(
	httpPort.NewServer,
	httpPort.NewHandler,
	httpPort.NewAccessContextProvider,
	remotes.NewRepository,
	wire.Bind(new(http.Handler), new(*httpPort.Handler)),
	wire.Bind(new(httpPort.AuthProvider), new(*httpPort.AccessContextProvider)),
	wire.Bind(new(httpPort.RemoteRepository), new(*remotes.Repository)),
)
