package jetstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/alpes214/stellar-hooks/internal/events"
)

type msgPublisher interface {
	PublishMsg(m *nats.Msg, opts ...nats.PubOpt) (*nats.PubAck, error)
}

type JetStreamProducer struct {
	pub     msgPublisher
	backoff time.Duration
}

func NewJetStreamProducer() *JetStreamProducer {
	return &JetStreamProducer{pub: JetStream, backoff: time.Second}
}

func (p *JetStreamProducer) PublishEvent(ctx context.Context, subject string, event *events.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set(nats.MsgIdHdr, event.ID)

	backoff := p.backoff
	if backoff == 0 {
		backoff = time.Second
	}

	for {
		ack, err := p.pub.PublishMsg(msg)
		if err == nil {
			log.Printf("Published event to JetStream: %s @ %d", subject, ack.Sequence)
			return nil
		}
		log.Printf("Failed to publish event to JetStream: %v", err)
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to publish event to subject %s: %w", subject, ctx.Err())
		case <-time.After(backoff):
		}
	}
}
