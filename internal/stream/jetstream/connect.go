package jetstream

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	NatsConn  *nats.Conn
	JetStream nats.JetStreamContext
)

func Connect(onClosed func()) error {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	opts := []nats.Option{
		nats.Name("Stellar Hooks"),
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(-1),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("Disconnected from NATS: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Reconnected to NATS at %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("Connection to NATS closed")
			if onClosed != nil {
				onClosed()
			}
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

	log.Println("Connected to NATS JetStream")

	cfg := &nats.StreamConfig{
		Name:       "EVENTS",
		Subjects:   []string{"stellar.events"},
		Storage:    nats.FileStorage,
		Retention:  nats.LimitsPolicy,
		MaxAge:     24 * time.Hour,
		MaxMsgs:    1_000_000,
		MaxBytes:   10 * 1024 * 1024 * 1024,
		Duplicates: time.Hour,
	}

	_, err = js.StreamInfo(cfg.Name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = js.AddStream(cfg)
	} else if err == nil {
		_, err = js.UpdateStream(cfg)
	}
	return err
}
