package controller

import (
	"encoding/binary"
	"encoding/json"

	bolt "github.com/coreos/bbolt"
	"github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"
)

func (b *BrokerController) CreateServiceInstance(ID string, req *v2.CreateServiceInstanceRequest) (*v2.CreateServiceInstanceResponse, int, error) {

	// create a record that a create service instance was requested.
	// save that record to the db.

	u := User{}

	b.db.Update(func(tx *bolt.Tx) error {
		// Retrieve the users bucket.
		// This should be created when the DB is first opened.
		buc := tx.Bucket([]byte("users"))

		// Generate ID for the user.
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := buc.NextSequence()
		u.ID = int(id)

		// Marshal user data into bytes.
		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		// Persist bytes to users bucket.
		return buc.Put(itob(u.ID), buf)
	})

	return nil, 0, nil
}

func (b *BrokerController) UpdateServiceInstance(ID string, req *v2.CreateServiceInstanceRequest) (*v2.ServiceInstance, int, error) {
	return nil, 0, nil
}

func (b *BrokerController) DeleteServiceInstance(ID string, req *v2.DeleteServiceInstanceRequest) (*v2.DeleteServiceInstanceResponse, int, error) {
	return nil, 0, nil
}

func (b *BrokerController) PollServiceInstance(ID string, req *v2.LastOperationRequest) (*v2.LastOperationResponse, int, error) {
	return nil, 0, nil
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

type User struct {
	ID int
}
