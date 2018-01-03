package controller

import (
	bolt "github.com/coreos/bbolt"
	p "github.com/n3wscott/k8s-broker-proxy/pkg/persistance"
	"github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"
)

func (b *BrokerController) GetCatalog() (*v2.Catalog, error) {

	b.db.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket([]byte(p.Metrics))
		err := buc.Put([]byte(p.Catalog), []byte("2")) // TODO this should inc.
		return err
	})

	catalog := v2.Catalog{}
	return &catalog, nil
}
