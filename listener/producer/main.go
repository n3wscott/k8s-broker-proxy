package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"flag"

	"os/signal"
	"syscall"

	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/golang/glog"
)

var options struct {
	ProjectID string
	Topic     string
}

func init() {
	flag.StringVar(&options.ProjectID, "projectId", "", "specify the gcp projectId")
	flag.StringVar(&options.Topic, "topic", "", "specify the pub/sub topic")
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
	topic  *pubsub.Topic
	client *pubsub.Client
)

func runWithContext(ctx context.Context) error {

	var err error
	client, err = pubsub.NewClient(ctx, options.ProjectID)
	if err != nil {
		log.Fatal(err)
	}

	topic = client.Topic(options.Topic)
	if _, err := topic.Exists(ctx); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", listHandler)
	http.HandleFunc("/pubsub/publish", publishHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

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

func listHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("Could not execute template: %v", err)
	}
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
