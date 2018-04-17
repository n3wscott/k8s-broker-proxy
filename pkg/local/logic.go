package local

import (
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/messages"
	"github.com/n3wscott/k8s-broker-proxy/pkg/cli"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"encoding/json"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func NewBusinessLogic(o cli.Options) (*BusinessLogic, error) {
	reg, err := messages.NewRegistry(o.ProjectID, o.Topic, o.Subscription)
	if err != nil {
		glog.Fatal(err)
	}

	config := osb.DefaultClientConfiguration()
	config.URL = o.BrokerUrl

	client, err := osb.NewClient(config)
	if err != nil {
		glog.Fatal(err)
		return nil, err
	}

	b := &BusinessLogic{
		async:  o.Async,
		reg:    reg,
		client: client,
	}

	b.RegisterSinks()

	return b, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool

	reg *messages.Registry

	client osb.Client
}

type ResponseBody struct {
	Response interface{} `json:"response"`
	Error    interface{} `json:"error"`
}

func (b *BusinessLogic) RegisterSinks() {
	glog.Info("RegisterSinks")
	b.reg.Sink("GetCatalog", b.sinkGetCatalog)
	b.reg.Sink("Provision", b.sinkProvision)
	b.reg.Sink("Deprovision", b.sinkDeprovision)
	b.reg.Sink("LastOperation", b.sinkLastOperation)
	b.reg.Sink("Bind", b.sinkBind)
	b.reg.Sink("Unbind", b.sinkUnbind)
	b.reg.Sink("Update", b.sinkUpdate)
}

func convertBodyTo(body interface{}, req interface{}) {
	if body != nil {
		if data, err := json.Marshal(body); err != nil {
			glog.Error(err)
		} else if err := json.Unmarshal(data, &req); err != nil {
			glog.Error(err)
		}
	}
}

func (b *BusinessLogic) sinkGetCatalog(id string, body interface{}) {
	glog.Info("sinkGetCatalog ", id)
	resp, err := b.GetCatalog(nil)
	b.reg.VentWith(id, "GetCatalog", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) sinkProvision(id string, body interface{}) {
	var request osb.ProvisionRequest
	convertBodyTo(body, request)

	resp, err := b.Provision(&request, nil)
	b.reg.VentWith(id, "Provision", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) sinkDeprovision(id string, body interface{}) {
	var request osb.DeprovisionRequest
	convertBodyTo(body, request)

	resp, err := b.Deprovision(&request, nil)
	b.reg.VentWith(id, "Deprovision", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) sinkLastOperation(id string, body interface{}) {
	var request osb.LastOperationRequest
	convertBodyTo(body, request)

	resp, err := b.LastOperation(&request, nil)
	b.reg.VentWith(id, "LastOperation", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) sinkBind(id string, body interface{}) {
	var request osb.BindRequest
	convertBodyTo(body, request)

	resp, err := b.Bind(&request, nil)
	b.reg.VentWith(id, "Bind", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) sinkUnbind(id string, body interface{}) {
	var request osb.UnbindRequest
	convertBodyTo(body, request)

	resp, err := b.Unbind(&request, nil)
	b.reg.VentWith(id, "Unbind", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) sinkUpdate(id string, body interface{}) {
	var request osb.UpdateInstanceRequest
	convertBodyTo(body, request)

	resp, err := b.Update(&request, nil)
	b.reg.VentWith(id, "Update", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) AdditionalRouting(router *mux.Router) {
	// TODO: could pass in the router to the registry and it can do the assigning internally.
}

var _ broker.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	resp, err := b.client.GetCatalog()

	if err != nil {
		glog.Error("GetCatalog failed with ", err)
		return nil, err
	}

	return &broker.CatalogResponse{
		CatalogResponse: *resp,
	}, err
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	resp, err := b.client.ProvisionInstance(request)

	if err != nil {
		glog.Error("ProvisionInstance failed with ", err)
		return nil, err
	}

	return &broker.ProvisionResponse{
		ProvisionResponse: *resp,
	}, err
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	resp, err := b.client.DeprovisionInstance(request)

	if err != nil {
		glog.Error("DeprovisionInstance failed with ", err)
		return nil, err
	}

	return &broker.DeprovisionResponse{
		DeprovisionResponse: *resp,
	}, err
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	resp, err := b.client.PollLastOperation(request)

	if err != nil {
		glog.Error("PollLastOperation failed with ", err)
		return nil, err
	}

	return &broker.LastOperationResponse{
		LastOperationResponse: *resp,
	}, err
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	resp, err := b.client.Bind(request)

	if err != nil {
		glog.Error("Bind failed with ", err)
		return nil, err
	}

	return &broker.BindResponse{
		BindResponse: *resp,
	}, err
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	resp, err := b.client.Unbind(request)

	if err != nil {
		glog.Error("Unbind failed with ", err)
		return nil, err
	}

	return &broker.UnbindResponse{
		UnbindResponse: *resp,
	}, err
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	resp, err := b.client.UpdateInstance(request)

	if err != nil {
		glog.Error("UpdateInstance failed with ", err)
		return nil, err
	}

	return &broker.UpdateInstanceResponse{
		UpdateInstanceResponse: *resp,
	}, err
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	glog.Info("ValidateBrokerAPIVersion")
	return nil
}
