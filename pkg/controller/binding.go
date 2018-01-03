package controller

import "github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"

func (b *BrokerController) CreateServiceBinding(instanceID, bindingID string, req *v2.BindingRequest) (*v2.CreateServiceBindingResponse, error) {
	return nil, nil
}

func (b *BrokerController) DeleteServiceBinding(instanceID, bindingID, serviceID, planID string) error {
	return nil
}
