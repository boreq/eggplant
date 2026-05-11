package adapters

import (
	"time"

	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

func NewBolt(path string) (*bolt.DB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return nil, errors.Wrap(err, "error opening the database (the database file is locked in exclusive mode, is another instance of the program running?)")
		}
		return nil, errors.Wrap(err, "error opening the database")
	}

	return db, nil
}
