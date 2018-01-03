package controller

import "github.com/n3wscott/osb-framework-go/pkg/apis/broker/v2"

func (b *BrokerController) CreateServiceInstance(ID string, req *v2.CreateServiceInstanceRequest) (*v2.CreateServiceInstanceResponse, int, error) {
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
