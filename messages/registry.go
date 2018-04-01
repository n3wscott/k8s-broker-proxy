package messages

import (
	"context"
	"time"

	"encoding/json"

	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/golang/glog"
	"github.com/pborman/uuid"
)

const DefaultWaitForTimeoutSec = 30
const DefaultSinkPollTimeSec = 5

// NewRegistry wraps a pub/sub topic and subscription.
func NewRegistry(projectID, topic, subscription string) (*Registry, error) {

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		glog.Fatal("failed to create client, ", err)
		return nil, err
	}

	clientTopic := client.Topic(topic)
	if _, err := clientTopic.Exists(ctx); err != nil {
		glog.Fatal("failed to verify topic exists: ", err)
		return nil, err
	}

	clientSubscription := client.Subscription(subscription)
	if ok, err := clientSubscription.Exists(ctx); err != nil || !ok {
		glog.Fatal("failed to create client subscription, exists: ", ok, " error: ", err)
		return nil, err
	}

	r := &Registry{
		WaitForTimeout: time.Second * DefaultWaitForTimeoutSec,
		SinkPollTime:   time.Second * DefaultSinkPollTimeSec,

		client:       client,
		topic:        clientTopic,
		subscription: clientSubscription,

		sinks: make(map[string]*Subscription, 10),
	}
	return r, nil
}

func (r *Registry) Vent(event string, body interface{}) (*string, error) {
	id := uuid.NewUUID().String()

	return r.vent(id, event, body)
}

func (r *Registry) VentWith(id, event string, body interface{}) (*string, error) {
	return r.vent(id, event, body)
}

func (r *Registry) vent(id, event string, body interface{}) (*string, error) {
	ctx := context.Background()

	json, err := json.Marshal(Message{
		ID:    id,
		Event: event,
		Body:  body,
	})
	if err != nil {
		glog.Errorf("failed to marshal body: %v", err)
		return nil, err
	}

	msg := &pubsub.Message{Data: json}

	if _, err := r.topic.Publish(ctx, msg).Get(ctx); err != nil {
		glog.Errorf("could not publish message: %v", err)
		return nil, err
	}

	glog.V(9).Info("message published")

	return &id, nil
}

func (r *Registry) SinkWorker() {
	ctx := context.Background()
	glog.Info("starting sink worker")

	r.sinking = true
	cctx, cancel := context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			glog.Info("worker quit")
			return
		case <-time.After(r.SinkPollTime):
			if len(r.sinks) == 0 {
				cancel()
				continue
			}
			err := r.subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
				r.sinkMutex.Lock()
				defer r.sinkMutex.Unlock()

				glog.Info("Got message: ", string(msg.Data))

				// TODO: I am not sure about this
				if msg.ID == "abort" {
					cancel()
				}

				message := &Message{}
				err := json.Unmarshal(msg.Data, message)
				if err != nil {
					glog.Error(err)
				}

				if s := r.sinks[message.ID]; s != nil {
					msg.Ack()
					go s.Callback(message.ID, message.Body)
				} else if s := r.sinks[message.Event]; s != nil {
					msg.Ack()
					go s.Callback(message.ID, message.Body)
				} else {
					msg.Nack()
				}
			})
			if err != nil {
				glog.Error(err)
				// TODO: cancel here too?
				return
			}
		}
	}
	r.sinking = false
}

// Skin will watch the subscription for messages with a matching event and call the callback with the body of the message.
func (r *Registry) Sink(key string, callback Callback) error {
	// TODO: add some more validation handling here.
	if r.sinks[key] != nil {
		return fmt.Errorf("error: sink exists for %s", key)
	}
	sink := &Subscription{
		Key:      key,
		Callback: callback,
	}
	r.sinks[key] = sink
	if !r.sinking {
		go r.SinkWorker()
	}
	return nil
}

func (r *Registry) RemoveSink(event string) {
	if r.sinks[event] != nil {
		delete(r.sinks, event)
	}
}

// WaitFor will block until a message arrives with the matching id provided
func (r *Registry) WaitFor(id string) (interface{}, error) {
	response := make(chan interface{})

	r.Sink(id, func(event string, body interface{}) {
		response <- body
	})

	var resp interface{}
	var err error
	select {
	case resp = <-response:
		glog.Info(id, " response received ", resp)
	case <-time.After(r.WaitForTimeout):
		glog.Error(id, " - timeout")
		err = fmt.Errorf("timeout")
	}

	r.RemoveSink(id)

	return resp, err
}

// topic string, subscription string

// Vent - needs client and topic
// WaitFor(Id) - needs client and subscription
// Sink - needs client and subscription

// For my case:
//
// Proxy  -A->  Local Proxy
// Proxy  <-B- Local Proxy

// - proxy vents on A and
// -- then waits on B for id or timeout
// - local proxy listens on A
// -- then local proxy vents on B

// Flow:

// Proxy           A             B         Local Proxy           Action
//  T -- event --> +                                             Vent  (T on A)
//  S ---------------- wait  --> ]                               WaitFor (S on B)
//                 S -------- Process ------> ]                  Listen (S on A)
//                               + <- Respond T                  Vent (T on B)
//  S <------ Response ----------]                               end WaitFor (S on B)

// Assumptions:
// - Vent is always going to happen on either A or B but never both for each side.
// - Sink does not need to know the topic it relates to in the API
//  scope creep: - WaitFor needs to be able to handle the case where WaitFor is called after the message is received.
