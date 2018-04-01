package local

import (
	"reflect"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/messages"
	"github.com/n3wscott/k8s-broker-proxy/pkg/cli"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

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
	b.reg.Sink("GetCatalog", b.ventGetCatalog)
	b.reg.Sink("Provision", b.ventProvision)
	b.reg.Sink("Deprovision", b.ventDeprovision)
	b.reg.Sink("LastOperation", b.ventLastOperation)
	b.reg.Sink("Bind", b.ventBind)
	b.reg.Sink("Unbind", b.ventUnbind)
	b.reg.Sink("Update", b.ventUpdate)
}

func (b *BusinessLogic) ventGetCatalog(id string, body interface{}) {
	resp, err := b.GetCatalog(nil)
	b.reg.VentWith(id, "GetCatalog", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) ventProvision(id string, body interface{}) {
	request := body.(*osb.ProvisionRequest)
	resp, err := b.Provision(request, nil)
	b.reg.VentWith(id, "Provision", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) ventDeprovision(id string, body interface{}) {
	request := body.(*osb.DeprovisionRequest)
	resp, err := b.Deprovision(request, nil)
	b.reg.VentWith(id, "Deprovision", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) ventLastOperation(id string, body interface{}) {
	request := body.(*osb.LastOperationRequest)
	resp, err := b.LastOperation(request, nil)
	b.reg.VentWith(id, "LastOperation", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) ventBind(id string, body interface{}) {
	request := body.(*osb.BindRequest)
	resp, err := b.Bind(request, nil)
	b.reg.VentWith(id, "Bind", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) ventUnbind(id string, body interface{}) {
	request := body.(*osb.UnbindRequest)
	resp, err := b.Unbind(request, nil)
	b.reg.VentWith(id, "Unbind", ResponseBody{
		Response: resp,
		Error:    err,
	})
}

func (b *BusinessLogic) ventUpdate(id string, body interface{}) {
	request := body.(*osb.UpdateInstanceRequest)
	resp, err := b.Update(request, nil)
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
