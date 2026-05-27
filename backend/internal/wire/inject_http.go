package wire

import (
	"net/http"

	httpPort "github.com/boreq/eggplant/entrypoints/http"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var httpSet = wire.NewSet(
	httpPort.NewServer,
	httpPort.NewHandler,
	httpPort.NewAccessContextProvider,
	wire.Bind(new(http.Handler), new(*httpPort.Handler)),
	wire.Bind(new(httpPort.AuthProvider), new(*httpPort.AccessContextProvider)),
)
