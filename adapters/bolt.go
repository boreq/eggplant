package adapters

import (
	"os"
	"path/filepath"
	"time"

	"github.com/boreq/errors"
	bolt "go.etcd.io/bbolt"
)

func NewBolt(path string) (*bolt.DB, error) {
	dir, _ := filepath.Split(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "mkdir all error")
	}

	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, errors.Wrap(err, "error opening the database")
	}

	return db, nil
}
