package jetstream

import (
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	NatsConn  *nats.Conn
	JetStream nats.JetStreamContext
)

func Connect() error {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	opts := []nats.Option{
		nats.Name("Stellar Hooks"),
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(10), // Retry up to 10 times
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("Disconnected from NATS: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Reconnected to NATS at %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("Connection to NATS closed")
		}),
	}

	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return err
	}
	NatsConn = nc

	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	JetStream = js

	log.Println("✅ Connected to NATS JetStream")

	_, err = js.AddStream(&nats.StreamConfig{
		Name:      "EVENTS",
		Subjects:  []string{"stellar.events"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return err
	}

	return nil
}

func InitJetStream() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	nc, err := nats.Connect(natsURL,
		nats.Name("Stellar Hooks"),
		nats.Timeout(10*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to initialize JetStream context: %v", err)
	}

	JetStream = js
	log.Println("✅ Connected to NATS JetStream")

	// Create a stream if it doesn’t exist
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      "EVENTS",
		Subjects:  []string{"stellar.events"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Fatalf("Failed to create stream: %v", err)
	}
}
