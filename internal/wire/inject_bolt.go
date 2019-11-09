package wire

import (
	"path/filepath"

	"github.com/boreq/eggplant/adapters"
	"github.com/boreq/eggplant/internal/config"
	"github.com/google/wire"
	bolt "go.etcd.io/bbolt"
)

//lint:ignore U1000 because
var boltSet = wire.NewSet(
	newBolt,
)

func newBolt(conf *config.Config) (*bolt.DB, error) {
	path := filepath.Join(conf.DataDirectory, "eggplant.database")
	return adapters.NewBolt(path)
}
