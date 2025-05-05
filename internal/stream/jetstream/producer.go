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

func GetLastCursorFromStream() (string, error) {
	streamName := "EVENTS"

	info, err := JetStream.StreamInfo(streamName)
	if err != nil {
		return "", fmt.Errorf("failed to get stream info: %w", err)
	}

	lastSeq := info.State.LastSeq
	if lastSeq == 0 {
		return "", nil // No events yet
	}

	msg, err := JetStream.GetMsg(streamName, lastSeq)
	if err != nil {
		return "", fmt.Errorf("failed to fetch last message: %w", err)
	}

	var evt events.Event
	if err := json.Unmarshal(msg.Data, &evt); err != nil {
		return "", fmt.Errorf("failed to unmarshal last event: %w", err)
	}

	return evt.ID, nil // Horizon's paging_token is stored as evt.ID
}
