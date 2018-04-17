package local

import (
	"reflect"

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

	b := &BusinessLogic{
		async: o.Async,
		reg:   reg,
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
	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "example-starter-pack-service",
				ID:            "4f6e6cf6-ffdd-425f-a2c7-3c9258ad246a",
				Description:   "The example service from the osb starter pack!",
				Bindable:      true,
				PlanUpdatable: truePtr(),
				Metadata: map[string]interface{}{
					"displayName": "Example starter pack service",
					"imageUrl":    "https://avatars2.githubusercontent.com/u/19862012?s=200&v=4",
				},
				Plans: []osb.Plan{
					{
						Name:        "default",
						ID:          "86064792-7ea2-467b-af93-ac9694d96d5b",
						Description: "The default plan for the starter pack example service",
						Free:        truePtr(),
						Schemas: &osb.Schemas{
							ServiceInstance: &osb.ServiceInstanceSchema{
								Create: &osb.InputParametersSchema{
									Parameters: map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"color": map[string]interface{}{
												"type":    "string",
												"default": "Clear",
												"enum": []string{
													"Clear",
													"Beige",
													"Grey",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	response.CatalogResponse = *osbResponse
	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	response := broker.ProvisionResponse{}
	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	response := broker.DeprovisionResponse{}

	//if request.AcceptsIncomplete {
	//	response.Async = b.async
	//}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {

	response := broker.LastOperationResponse{}

	// Your last-operation business logic goes here

	return &response, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	//if request.AcceptsIncomplete {
	//	response.Async = b.async
	//}

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: map[string]interface{}{"todo": "todo"},
		},
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	glog.Info("ValidateBrokerAPIVersion")
	return nil
}

type Instance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
}

func (i *Instance) Match(other *Instance) bool {
	return reflect.DeepEqual(i, other)
}
