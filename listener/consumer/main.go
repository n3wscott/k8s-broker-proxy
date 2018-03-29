package main

import (
	"context"
	"log"
	"os"

	"flag"

	"os/signal"
	"syscall"

	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/golang/glog"
)

var options struct {
	ProjectID    string
	Subscription string
}

const maxMessages = 100

func init() {
	flag.StringVar(&options.ProjectID, "projectId", "", "specify the gcp projectId")
	flag.StringVar(&options.Subscription, "subscription", "", "specify the pub/sub subscription")
	flag.Parse()
}

func main() {
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		glog.Fatalln("Failed with: ", err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

var (
	subscription *pubsub.Subscription

	client *pubsub.Client
)

func runWithContext(ctx context.Context) error {

	var err error
	client, err = pubsub.NewClient(ctx, options.ProjectID)
	if err != nil {
		log.Fatal(err)
	}

	subscription = client.Subscription(options.Subscription)
	if ok, err := subscription.Exists(ctx); err != nil || !ok {
		glog.Info("Exists: ", ok, " Error: ", err)
	}

	// Consume 10 messages.
	var mu sync.Mutex
	received := 0

	glog.Info("going to start listening for messages")

	cctx, cancel := context.WithCancel(ctx)
	err = subscription.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		mu.Lock()
		defer mu.Unlock()
		received++
		if received >= maxMessages {
			cancel()
			msg.Nack()
			return
		}
		glog.Info("Got message: ", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		return err
	}

	glog.Info("done listening for messages")
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
