package producer

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/n3wscott/k8s-broker-proxy/pkg/cli"
	"github.com/pborman/uuid"
	"github.com/pmorie/osb-broker-lib/pkg/broker"

	"context"

	"time"

	"github.com/gin-gonic/gin/json"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

const timeoutSeconds = 30
const workerPollSeconds = 3

// The work flow for this application is as follows:
// 1. broker is acting as a broker accepting requests locally
// 2. broker takes each requests and publishes it to the broker-proxy-to-broker queue
// 3. broker waits for a response from the broker-to-broker-proxy queue tied to request and responds when that is received.

// In another location:
// 1. broker is waiting for requests to come in over the queue.
// 2. broker unpacks the request and composes a OSB call to the configured local broker.
// 3. broker packs the response and replies on the broker-to-broker-proxy queue

// Outstanding questions:
//  - how do we do auth on this?
//  > Maybe we OOB the auth between the ends and leave it as a todo

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o cli.Options) (*BusinessLogic, error) {
	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, o.ProjectID)
	if err != nil {
		glog.Fatal(err)
	}

	topic := client.Topic(o.Topic)
	if _, err := topic.Exists(ctx); err != nil {
		glog.Fatal(err)
	}

	b := &BusinessLogic{
		async:     o.Async,
		instances: make(map[string]*Instance, 10),
		pending:   make(map[string]*PendingRequest, 10),
		client:    client,
		topic:     topic,
	}

	go b.worker()

	return b, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// todo
	instances map[string]*Instance

	catalog *broker.CatalogResponse

	// Pubsub
	client *pubsub.Client
	topic  *pubsub.Topic

	pending map[string]*PendingRequest
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

type PendingRequest struct {
	ID       string
	Done     chan<- bool
	Response chan PendingResponse
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
	for {
		select {
		case <-ctx.Done():
			glog.Info("worker quit")
			return
		case <-time.After(workerPollSeconds * time.Second):
			// do work.
			for _, r := range b.pending {
				glog.Info("working on - ", r)
				sample := fmt.Sprintf("{\"id\":\"%s\"}", r.ID)
				r.Response <- PendingResponse{
					ID:       r.ID,
					Method:   "Demo",
					Response: []byte(sample),
				}
			}
		}
	}
}

func (b *BusinessLogic) waitFor(id string) (resp PendingResponse, err error) {
	// Start a worker goroutine, giving it the channel to
	// notify on.
	done := make(chan bool)
	response := make(chan PendingResponse)

	b.pending[id] = &PendingRequest{
		ID:       id,
		Response: response,
		Done:     done,
	}

	select {
	case resp = <-response:
		glog.Info(id, " response received ", resp)
	case <-done:
		glog.Info(id, " - is done")
		err = fmt.Errorf("done")
	case <-time.After(timeoutSeconds * time.Second):
		glog.Info(id, " - timeout")
		err = fmt.Errorf("timeout")
	}

	glog.Info(id, " - clearing")
	// Remember to clear out the pending request.
	delete(b.pending, id)

	return
}

// TODO: might want to return a context object rather than string for id.
func (b *BusinessLogic) publish(method string, request interface{}) (string, error) {
	ctx := context.Background()

	id := uuid.NewUUID().String()
	json, err := json.Marshal(PubSubData{
		ID:      id,
		Method:  method,
		Request: request,
	})
	if err != nil {
		glog.Error(err)
	}

	msg := &pubsub.Message{
		Data: json,
		// TODO this needs the method and whatnot
		// and some kind of id...
	}

	if _, err := b.topic.Publish(ctx, msg).Get(ctx); err != nil {
		glog.Errorf("Could not publish message: %v", err)
		return "", err
	}

	glog.Info("Message published.")

	// TODO: next we should create a chan and wait
	// for a timeout or that chan to report a response back.

	return id, nil
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	id, err := b.publish("GetCatalog", nil)
	if err != nil {
		// todo
	}
	b.waitFor(id)

	if b.catalog != nil {
		return b.catalog, nil
	}

	// Your catalog business logic goes here
	response := &broker.CatalogResponse{}

	// save the catalog for later
	b.catalog = response
	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	id, _ := b.publish("Provision", request)
	b.waitFor(id)

	// Your provision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}

	instance := &Instance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
	}

	if b.instances[request.InstanceID] != nil {
		i := b.instances[request.InstanceID]
		if i.Match(instance) {
			response.Exists = true
		} else {
			glog.Error("InstanceID in use")
			return nil, fmt.Errorf("InstanceID in use")
		}
	}

	b.instances[request.InstanceID] = instance

	// when we support async:
	//if request.AcceptsIncomplete {
	//	response.Async = b.async
	//}

	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	b.publish("Deprovision", request)

	// Your deprovision business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}

	// TODO

	delete(b.instances, request.InstanceID)

	//if request.AcceptsIncomplete {
	//	response.Async = b.async
	//}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	b.publish("LastOperation", request)

	// Your last-operation business logic goes here

	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	b.publish("Bind", request)

	// Your bind business logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	_, ok := b.instances[request.InstanceID]
	if !ok {
		return nil, osb.HTTPStatusCodeError{
			StatusCode: http.StatusNotFound,
		}
	}

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
	b.publish("Unbind", request)

	// Your unbind business logic goes here
	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	b.publish("UpdateInstance", request)

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
