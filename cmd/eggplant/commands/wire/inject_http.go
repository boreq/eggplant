package wire

import (
	"net/http"

	httpPort "github.com/boreq/eggplant/pkg/service/ports/http"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var httpSet = wire.NewSet(
	httpPort.NewServer,
	httpPort.NewHandler,
	wire.Bind(new(http.Handler), new(*httpPort.Handler)),
)
