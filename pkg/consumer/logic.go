package consumer

import (
	"reflect"

	"cloud.google.com/go/pubsub"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/pkg/cli"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"context"

	"time"

	"encoding/json"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

func NewBusinessLogic(o cli.Options) (*BusinessLogic, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, o.ProjectID)
	if err != nil {
		glog.Fatal(err)
	}

	topic := client.Topic(o.Topic)
	if _, err := topic.Exists(ctx); err != nil {
		glog.Fatal(err)
	}

	subscription := client.Subscription(o.Subscription)
	if ok, err := subscription.Exists(ctx); err != nil || !ok {
		glog.Fatal("Subscription Exists: ", ok, " Error: ", err)
	}

	b := &BusinessLogic{
		async:        o.Async,
		client:       client,
		topic:        topic,
		subscription: subscription,
	}

	go b.worker()

	return b, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool

	// Pubsub
	client       *pubsub.Client
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
}

func (b *BusinessLogic) AdditionalRouting(router *mux.Router) {
	// TODO: could pass in the router to the registry and it can do the assigning internally.
}

var _ broker.Interface = &BusinessLogic{}

// todo: might want to break this apart out of the logic area
type PubSubData struct {
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Request interface{} `json:"request"`
}

type PendingResponse struct {
	ID       string      `json:"id"`
	Method   string      `json:"method"`
	Response interface{} `json:"response"`
}

// for now we will just hang for 5 seconds.
func (b *BusinessLogic) worker() {
	ctx := context.Background()
	glog.Info("starting worker")
	cctx, cancel := context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			glog.Info("worker quit")
			return
		case <-time.After(time.Second):
			// do work.
			err := b.subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
				//mu.Lock()
				//defer mu.Unlock()
				glog.Info("Got message: ", string(msg.Data))

				if msg.ID == "abort" {
					cancel()
				}
				request := &PubSubData{}
				err := json.Unmarshal(msg.Data, request)
				if err != nil {
					glog.Error(err)
				}
				msg.Ack()

				go b.process(request)
			})
			if err != nil {
				glog.Error(err)
				return
			}
		}
	}
}

func (b *BusinessLogic) process(req *PubSubData) {

	switch req.Method {
	case "GetCatalog":
		resp, err := b.GetCatalog(nil)
		if err != nil {
			glog.Error(err)
		}
		b.publish(req.ID, req.Method, resp)
	case "Provision":
		provisionRequest := req.Request.(*osb.ProvisionRequest)
		resp, err := b.Provision(provisionRequest, nil)
		if err != nil {
			glog.Error(err)
		}
		b.publish(req.ID, req.Method, resp)
	case "Deprovision":
	case "LastOperation":
	case "Bind":
	case "Unbind":
	case "Update":
	}
}

// TODO: might want to return a context object rather than string for id.
func (b *BusinessLogic) publish(id, method string, response interface{}) (string, error) {
	ctx := context.Background()

	json, err := json.Marshal(PendingResponse{
		ID:       id,
		Method:   method,
		Response: response,
	})
	if err != nil {
		glog.Error(err)
	}

	msg := &pubsub.Message{
		Data: json,
	}

	if _, err := b.topic.Publish(ctx, msg).Get(ctx); err != nil {
		glog.Errorf("Could not publish message: %v", err)
		return "", err
	}

	glog.Info("Message published.")

	return id, nil
}

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
