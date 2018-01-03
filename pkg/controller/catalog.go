package controller

import (
	"github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"
)

func (b *BrokerController) GetCatalog() (*v2.Catalog, error) {
	catalog := v2.Catalog{}
	return &catalog, nil
}
