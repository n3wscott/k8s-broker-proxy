package proxy

import (
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/pkg/cli"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"github.com/n3wscott/k8s-broker-proxy/messages"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func NewBusinessLogic(o cli.Options) (*BusinessLogic, error) {

	reg, err := messages.NewRegistry(o.ProjectID, o.Topic, o.Subscription)
	if err != nil {
		glog.Fatal(err)
	}

	b := &BusinessLogic{
		async: o.Async,
		reg:   reg,
	}

	return b, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool

	//	catalog *broker.CatalogResponse

	reg *messages.Registry
}

func (b *BusinessLogic) AdditionalRouting(router *mux.Router) {
	// TODO: could pass in the router to the registry and it can do the assigning internally.
}

var _ broker.Interface = &BusinessLogic{}

type ResponseBody struct {
	Response interface{} `json:"response"`
	Error    interface{} `json:"error"`
}

func (b *BusinessLogic) ventAndWait(method string, request interface{}) (*ResponseBody, error) {
	id, err := b.reg.Vent(method, request)

	if err != nil {
		return nil, err
	}

	body, err := b.reg.WaitFor(*id)
	if err != nil {
		//  pubsub issue or timeout?
	}
	// body is a message response, with an error and a response.

	resp := body.(ResponseBody)
	return &resp, nil
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	resp, err := b.ventAndWait("GetCatalog", nil)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.CatalogResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	resp, err := b.ventAndWait("Provision", request)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.ProvisionResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	resp, err := b.ventAndWait("Deprovision", request)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.DeprovisionResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	resp, err := b.ventAndWait("LastOperation", request)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.LastOperationResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	resp, err := b.ventAndWait("Bind", request)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.BindResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	resp, err := b.ventAndWait("Unbind", request)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.UnbindResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	resp, err := b.ventAndWait("Update", request)
	if err != nil {
		// todo?
		return nil, err
	}

	remoteResponse := resp.Response.(broker.UpdateInstanceResponse)
	remoteErr := resp.Error.(error)
	return &remoteResponse, remoteErr
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	glog.Info("ValidateBrokerAPIVersion")
	return nil
}
