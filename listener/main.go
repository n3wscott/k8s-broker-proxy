package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"flag"

	"os/signal"
	"syscall"

	"time"

	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/golang/glog"
	"google.golang.org/api/iterator"
)

var options struct {
	ProjectID    string
	Topic        string
	Subscription string
}

const maxMessages = 10

func init() {
	flag.StringVar(&options.ProjectID, "projectId", "", "specify the gcp projectId")
	flag.StringVar(&options.Topic, "topic", "", "specify the pub/sub topic")
	flag.StringVar(&options.Subscription, "subscription", "", "specify the pub/sub subscription")
	flag.Parse()
}

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		glog.Fatalln(err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

var (
	topic *pubsub.Topic

	// Messages received by this instance.
	messagesMu sync.Mutex
	messages   []string

	client *pubsub.Client
)

func runWithContext(ctx context.Context) error {

	var err error
	client, err = pubsub.NewClient(ctx, options.ProjectID) //mustGetenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}

	// Create topic if it doesn't exist.
	topic = client.Topic(options.Topic)
	if _, err := topic.Exists(ctx); err != nil {
		log.Fatal(err)
	}
	//topic, err = client.CreateTopic(ctx, options.Topic) //mustGetenv("PUBSUB_TOPIC"))
	//if err != nil {
	//	log.Fatal(err)
	//}

	http.HandleFunc("/", listHandler)
	http.HandleFunc("/pubsub/publish", publishHandler)
	http.HandleFunc("/pubsub/push", pushHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

	//sub, err := client.CreateSubscription(ctx, "lister-demo-subscription", pubsub.SubscriptionConfig{
	//	Topic: topic,
	//})
	//if err != nil {
	//	log.Fatal(err)
	//}

	//
	//// [START auth]
	////proj := options.ProjectID
	////if proj == "" {
	////	fmt.Fprintf(os.Stderr, "--projectId must be se\n")
	////	os.Exit(1)
	////}
	//proj := os.Getenv("GOOGLE_CLOUD_PROJECT")
	//if proj == "" {
	//	fmt.Fprintf(os.Stderr, "GOOGLE_CLOUD_PROJECT environment variable must be set.\n")
	//	os.Exit(1)
	//}
	//client, err := pubsub.NewClient(ctx, proj)
	//if err != nil {
	//	log.Fatalf("Could not create pubsub Client: %v", err)
	//}
	//// [END auth]
	//
	//// Print all the subscriptions in the project.
	//fmt.Println("Listing all subscriptions from the project:")
	//subs, err := list(client)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, sub := range subs {
	//	fmt.Println(sub)
	//}
	//
	//sub := options.Subscription
	//topic := options.Topic
	//
	//t := client.Topic(topic)
	//if _, err := t.Exists(ctx); err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Create a new subscription.
	//if err := create(client, sub, t); err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Pull messages via the subscription.
	//if err := pullMsgs(client, sub, t); err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Delete the subscription.
	//if err := delete(client, sub); err != nil {
	//	log.Fatal(err)
	//}

	return nil
}

func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			glog.Infof("Received SIGTERM, exiting gracefully...")
			f()
			os.Exit(0)
		case <-ctx.Done():
			os.Exit(0)
		}
	}
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	return v
}

type pushRequest struct {
	Message struct {
		Attributes map[string]string
		Data       []byte
		ID         string `json:"message_id"`
	}
	Subscription string
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	msg := &pushRequest{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, fmt.Sprintf("Could not decode body: %v", err), http.StatusBadRequest)
		return
	}

	messagesMu.Lock()
	defer messagesMu.Unlock()
	// Limit to ten.
	messages = append(messages, string(msg.Message.Data))
	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	messagesMu.Lock()
	defer messagesMu.Unlock()

	if err := tmpl.Execute(w, messages); err != nil {
		log.Printf("Could not execute template: %v", err)
	}

	// Create a new subscription.
	if err := create(client, options.Subscription, topic); err != nil {
		log.Fatal(err)
	}

	// Pull messages via the subscription.
	if err := pullMsgs(client, options.Subscription, topic); err != nil {
		log.Fatal(err)
	}
	//
	//// Delete the subscription.
	//if err := delete(client, options.Subscription); err != nil {
	//	log.Fatal(err)
	//}
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	msg := &pubsub.Message{
		Data: []byte(r.FormValue("payload")),
	}

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Could not publish message: %v", err), 500)
		return
	}

	fmt.Fprint(w, "Message published.")
}

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
  <head>
    <title>Pub/Sub</title>
  </head>
  <body>
    <div>
      <p>Last ten messages received by this instance:</p>
      <ul>
      {{ range . }}
          <li>{{ . }}</li>
      {{ end }}
      </ul>
    </div>
    <!-- [START form] -->
    <form method="post" action="/pubsub/publish">
      <textarea name="payload" placeholder="Enter message here"></textarea>
      <input type="submit">
    </form>
    <!-- [END form] -->
    <p>Note: if the application is running across multiple instances, each
      instance will have its own list of messages.</p>
  </body>
</html>`))

func list(client *pubsub.Client) ([]*pubsub.Subscription, error) {
	ctx := context.Background()
	// [START get_all_subscriptions]
	var subs []*pubsub.Subscription
	it := client.Subscriptions(ctx)
	for {
		s, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	// [END get_all_subscriptions]
	return subs, nil
}

func create(client *pubsub.Client, name string, topic *pubsub.Topic) error {
	ctx := context.Background()
	// [START create_subscription]
	sub, err := client.CreateSubscription(ctx, name, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Created subscription: %v\n", sub)
	// [END create_subscription]
	return nil
}

func delete(client *pubsub.Client, name string) error {
	ctx := context.Background()
	// [START delete_subscription]
	sub := client.Subscription(name)
	if err := sub.Delete(ctx); err != nil {
		return err
	}
	fmt.Println("Subscription deleted.")
	// [END delete_subscription]
	return nil
}

func pullMsgs(client *pubsub.Client, name string, topic *pubsub.Topic) error {
	ctx := context.Background()

	// Publish 10 messages on the topic.
	var results []*pubsub.PublishResult
	for i := 0; i < 10; i++ {
		res := topic.Publish(ctx, &pubsub.Message{
			Data: []byte(fmt.Sprintf("hello world #%d", i)),
		})
		results = append(results, res)
	}

	// Check that all messages were published.
	for _, r := range results {
		_, err := r.Get(ctx)
		if err != nil {
			return err
		}
	}

	// [START pull_messages]
	// Consume 10 messages.
	var mu sync.Mutex
	received := 0
	sub := client.Subscription(name)
	cctx, cancel := context.WithCancel(ctx)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()
		received++
		if received >= 10 {
			cancel()
			msg.Nack()
			return
		}
		fmt.Printf("Got message: %q\n", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		return err
	}
	// [END pull_messages]
	return nil
}
