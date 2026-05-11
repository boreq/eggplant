package fixture

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

type CleanupFunc func()

func File(t *testing.T) (string, CleanupFunc) {
	file, err := ioutil.TempFile("", "eggplant_test")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}

	return file.Name(), cleanup
}

func Bolt(t *testing.T) (*bolt.DB, CleanupFunc) {
	file, fileCleanup := File(t)

	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		defer fileCleanup()

		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	return db, cleanup
}
