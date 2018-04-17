package cli

import (
	"flag"
)

// Options holds the options specified by the broker's code on the command
// line. Users should add their own options here and add flags for them in
// AddFlags.
type Options struct {
	CatalogPath string
	Async       bool

	// For pubsub:
	ProjectID    string
	Topic        string
	Subscription string

	BrokerUrl string
}

// AddFlags is a hook called to initialize the CLI flags for broker options.
// It is called after the flags are added for the skeleton and before flag
// parse is called.
func AddFlags(o *Options) {
	flag.StringVar(&o.CatalogPath, "catalogPath", "", "The path to the catalog")
	flag.BoolVar(&o.Async, "async", false, "Indicates whether the broker is handling the requests asynchronously.")

	flag.StringVar(&o.ProjectID, "projectId", "", "specify the gcp projectId")
	flag.StringVar(&o.Topic, "topic", "", "specify the pub/sub topic")
	flag.StringVar(&o.Subscription, "subscription", "", "specify the pub/sub subscription")

	flag.StringVar(&o.BrokerUrl, "broker", "", "URL of the local broker")
}
