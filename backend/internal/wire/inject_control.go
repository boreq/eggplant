package wire

import (
	remotecontrolAdapters "github.com/boreq/eggplant/adapters/remotecontrol"
	"github.com/boreq/eggplant/application/remotecontrol"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var controlSet = wire.NewSet(
	wire.Struct(new(remotecontrol.Control), "*"),
	remotecontrol.NewConnectHandler,
	remotecontrol.NewDisconnectHandler,
	remotecontrol.NewSetControllableHandler,
	remotecontrol.NewRenameHandler,
	remotecontrol.NewTakeControlHandler,
	remotecontrol.NewReleaseControlHandler,
	remotecontrol.NewDisconnectControllersHandler,
	remotecontrol.NewSendCommandHandler,
	remotecontrol.NewPublishStateHandler,

	remotecontrolAdapters.NewHub,
	wire.Bind(new(remotecontrol.Hub), new(*remotecontrolAdapters.Hub)),
)
