package wire

import (
	"github.com/boreq/eggplant/entrypoints/filesystem"
	"github.com/google/wire"
)

//lint:ignore U1000 because
var filesystemSet = wire.NewSet(
	filesystem.NewListener,
)
