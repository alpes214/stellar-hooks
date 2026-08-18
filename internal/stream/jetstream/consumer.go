package jetstream

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/alpes214/stellar-hooks/internal/delivery"
	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/filter"
	"github.com/alpes214/stellar-hooks/internal/models"
)

const nakDelay = 2 * time.Second

type subscriptionLister interface {
	GetAllSubscriptions() ([]models.Subscription, error)
}

type jsMsg interface {
	Data() []byte
	Ack() error
	Term() error
	NakWithDelay(time.Duration) error
}

type natsMsg struct{ *nats.Msg }

func (m natsMsg) Data() []byte                       { return m.Msg.Data }
func (m natsMsg) Ack() error                         { return m.Msg.Ack() }
func (m natsMsg) Term() error                        { return m.Msg.Term() }
func (m natsMsg) NakWithDelay(d time.Duration) error { return m.Msg.NakWithDelay(d) }

type JetStreamConsumer struct {
	Subject string
	Durable string
	Store   subscriptionLister
	wg      sync.WaitGroup
}

func NewJetStreamConsumer(subject, durable string, store subscriptionLister) *JetStreamConsumer {
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

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.Println("JetStream consumer stopped")
				return
			default:
				msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
				if err != nil {
					if errors.Is(err, nats.ErrTimeout) {
						continue
					}
					log.Printf("Fetch error: %v", err)
					select {
					case <-ctx.Done():
						return
					case <-time.After(nakDelay):
					}
					continue
				}

				for _, msg := range msgs {
					c.process(natsMsg{msg}, delivery.SendToWebhook)
				}
			}
		}
	}()

	return nil
}

func (c *JetStreamConsumer) Wait() {
	c.wg.Wait()
}

func (c *JetStreamConsumer) process(msg jsMsg, send func(models.Subscription, *events.Event) error) {
	var evt events.Event
	if err := json.Unmarshal(msg.Data(), &evt); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		_ = msg.Term()
		return
	}

	subs, err := c.Store.GetAllSubscriptions()
	if err != nil {
		log.Printf("Error retrieving subscriptions: %v", err)
		_ = msg.NakWithDelay(nakDelay)
		return
	}

	var failed bool
	for _, sub := range subs {
		if filter.Matches(&sub, &evt) {
			log.Printf("Matched subscription ID=%d for event: %+v", sub.ID, evt)
			if err := send(sub, &evt); err != nil {
				log.Printf("Webhook delivery failed for subscription ID=%d: %v", sub.ID, err)
				failed = true
			}
		}
	}

	if failed {
		_ = msg.NakWithDelay(nakDelay)
		return
	}
	_ = msg.Ack()
}
