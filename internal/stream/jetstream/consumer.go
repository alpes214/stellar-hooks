package jetstream

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/alpes214/stellar-hooks/internal/delivery"
	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/filter"
	"github.com/alpes214/stellar-hooks/internal/storage"
)

type JetStreamConsumer struct {
	Subject string
	Durable string
	Store   storage.SubscriptionStore
}

func NewJetStreamConsumer(subject, durable string, store storage.SubscriptionStore) *JetStreamConsumer {
	return &JetStreamConsumer{
		Subject: subject,
		Durable: durable,
		Store:   store,
	}
}

func (c *JetStreamConsumer) Start(ctx context.Context) error {
	sub, err := JetStream.PullSubscribe(
		c.Subject,
		c.Durable,
		nats.PullMaxWaiting(128),
	)
	if err != nil {
		return err
	}

	log.Printf("Subscribed to JetStream subject: %s (durable: %s)", c.Subject, c.Durable)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("JetStream consumer stopped")
				return
			default:
				msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
				if err != nil && err != nats.ErrTimeout {
					log.Printf("Fetch error: %v", err)
					continue
				}

				for _, msg := range msgs {
					var evt events.Event
					if err := json.Unmarshal(msg.Data, &evt); err != nil {
						log.Printf("Failed to unmarshal event: %v", err)
						_ = msg.Nak()
						continue
					}

					subs, err := c.Store.GetAllSubscriptions()
					if err != nil {
						log.Printf("Error retrieving subscriptions: %v", err)
						_ = msg.Nak()
						continue
					}

					for _, sub := range subs {
						if filter.Matches(&sub, &evt) {
							log.Printf("Matched subscription ID=%d for event: %+v", sub.ID, evt)
							go delivery.SendToWebhook(sub, &evt)
						}
					}

					_ = msg.Ack()
				}
			}
		}
	}()

	return nil
}
