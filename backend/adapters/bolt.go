package adapters

import (
	"time"

	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
	bbolterrors "go.etcd.io/bbolt/errors"
)

func NewBolt(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		if errors.Is(err, bbolterrors.ErrTimeout) {
			return nil, errors.Wrap(err, "error opening the database (the database file is locked in exclusive mode, is another instance of the program running?)")
		}
		return nil, errors.Wrap(err, "error opening the database")
	}

	return db, nil
}
