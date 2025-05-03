package jetstream

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/alpes214/stellar-hooks/internal/events"
)

type JetStreamProducer struct{}

func NewJetStreamProducer() *JetStreamProducer {
	return &JetStreamProducer{}
}

func (p *JetStreamProducer) PublishEvent(subject string, event *events.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ack, err := JetStream.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish event to subject %s: %w", subject, err)
	}

	log.Printf("Published event to JetStream: %s @ %d", subject, ack.Sequence)
	return nil
}
