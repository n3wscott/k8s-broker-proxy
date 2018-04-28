package proxy

import (
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/pkg/binding"
	"github.com/n3wscott/k8s-broker-proxy/pkg/cli"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"encoding/json"

	"github.com/n3wscott/k8s-broker-proxy/messages"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func NewBusinessLogic(o cli.Options) (*BusinessLogic, error) {

	if o.Binding != "" {
		projectId, topicId, subscriptionId, err := binding.PubSubBinding(o.Binding)
		if err != nil {
			return nil, err
		}
		o.ProjectID = projectId
		o.Topic = topicId
		o.Subscription = subscriptionId
	}

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
		return nil, err
	}
	// body is a message response, with an error and a response.

	var resp ResponseBody

	if data, err := json.Marshal(body); err != nil {
		glog.Error(err)
	} else if err := json.Unmarshal(data, &resp); err != nil {
		glog.Error(err)
	}

	return &resp, nil
}

func convertRemoteResponse(resp *ResponseBody, remoteResponse interface{}, remoteErr error) {
	if resp.Response != nil {
		if data, err := json.Marshal(resp.Response); err != nil {
			glog.Error(err)
		} else if err := json.Unmarshal(data, &remoteResponse); err != nil {
			glog.Error(err)
		}
	}

	if resp.Error != nil {
		remoteErr = resp.Error.(error)
	}
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (remoteResponse *broker.CatalogResponse, remoteErr error) {
	resp, err := b.ventAndWait("GetCatalog", nil)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (remoteResponse *broker.ProvisionResponse, remoteErr error) {
	resp, err := b.ventAndWait("Provision", request)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (remoteResponse *broker.DeprovisionResponse, remoteErr error) {
	resp, err := b.ventAndWait("Deprovision", request)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (remoteResponse *broker.LastOperationResponse, remoteErr error) {
	resp, err := b.ventAndWait("LastOperation", request)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (remoteResponse *broker.BindResponse, remoteErr error) {
	resp, err := b.ventAndWait("Bind", request)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (remoteResponse *broker.UnbindResponse, remoteErr error) {
	resp, err := b.ventAndWait("Unbind", request)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (remoteResponse *broker.UpdateInstanceResponse, remoteErr error) {
	resp, err := b.ventAndWait("Update", request)
	if err != nil {
		return nil, err
	}

	convertRemoteResponse(resp, &remoteResponse, remoteErr)
	return
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	glog.Info("ValidateBrokerAPIVersion")
	return nil
}
