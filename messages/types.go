package messages

import (
	"time"

	"sync"

	"cloud.google.com/go/pubsub"
)

type Registry struct {
	WaitForTimeout time.Duration
	SinkPollTime   time.Duration

	client       *pubsub.Client
	topic        *pubsub.Topic
	subscription *pubsub.Subscription

	sinks     map[string]*Subscription // mapping the Key to a Subscription
	sinking   bool
	sinkMutex sync.Mutex
}

type Callback func(id string, body interface{})

type Subscription struct {
	Key      string
	Callback Callback
}

type Message struct {
	ID    string      `json:"id"`
	Event string      `json:"event"`
	Body  interface{} `json:"body"`
}
