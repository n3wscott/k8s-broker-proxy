package persistance

import (
	"fmt"

	bolt "github.com/coreos/bbolt"
)

const (
	Metrics = "Metrics"

	Catalog  = "Catalog"
	Instance = "Instance"
)

var db *bolt.DB

func SetupBolt(d *bolt.DB) {
	db = d
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(Metrics))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put([]byte(Catalog), []byte("1"))
		return err
	})
}
