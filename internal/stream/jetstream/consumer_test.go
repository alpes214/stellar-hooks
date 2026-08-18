package jetstream

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/models"
)

type fakeMsg struct {
	data     []byte
	acked    bool
	termed   bool
	nakDelay time.Duration
}

func (m *fakeMsg) Data() []byte                         { return m.data }
func (m *fakeMsg) Ack() error                           { m.acked = true; return nil }
func (m *fakeMsg) Term() error                          { m.termed = true; return nil }
func (m *fakeMsg) NakWithDelay(d time.Duration) error   { m.nakDelay = d; return nil }

type fakeSubs struct {
	subs []models.Subscription
	err  error
}

func (f *fakeSubs) GetAllSubscriptions() ([]models.Subscription, error) {
	return f.subs, f.err
}

func paymentJSON(t *testing.T) []byte {
	t.Helper()
	b, err := json.Marshal(&events.Event{ID: "op-1", Type: events.EventPayment})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestProcessTermsBadJSON(t *testing.T) {
	msg := &fakeMsg{data: []byte("not-json")}
	c := NewJetStreamConsumer("stellar.events", "webhook-subscriber", &fakeSubs{})

	c.process(msg, func(models.Subscription, *events.Event) error { return nil })

	if !msg.termed || msg.acked {
		t.Fatalf("termed=%v acked=%v", msg.termed, msg.acked)
	}
}

func TestProcessNaksStoreError(t *testing.T) {
	msg := &fakeMsg{data: paymentJSON(t)}
	c := NewJetStreamConsumer("stellar.events", "webhook-subscriber", &fakeSubs{err: errors.New("db")})

	c.process(msg, func(models.Subscription, *events.Event) error { return nil })

	if msg.nakDelay != nakDelay || msg.acked {
		t.Fatalf("nak=%v acked=%v", msg.nakDelay, msg.acked)
	}
}

func TestProcessAcksWhenNoMatch(t *testing.T) {
	msg := &fakeMsg{data: paymentJSON(t)}
	c := NewJetStreamConsumer("stellar.events", "webhook-subscriber", &fakeSubs{
		subs: []models.Subscription{{ID: 1, Types: []string{"account_created"}, WebhookURL: "http://x"}},
	})

	c.process(msg, func(models.Subscription, *events.Event) error {
		t.Fatal("should not send")
		return nil
	})

	if !msg.acked {
		t.Fatal("expected ack")
	}
}

func TestProcessAcksAfterSuccessfulDelivery(t *testing.T) {
	msg := &fakeMsg{data: paymentJSON(t)}
	c := NewJetStreamConsumer("stellar.events", "webhook-subscriber", &fakeSubs{
		subs: []models.Subscription{{ID: 1, Types: []string{"payment"}, WebhookURL: "http://x"}},
	})
	var sent int
	c.process(msg, func(models.Subscription, *events.Event) error {
		sent++
		return nil
	})

	if sent != 1 || !msg.acked {
		t.Fatalf("sent=%d acked=%v", sent, msg.acked)
	}
}

func TestProcessNaksWhenDeliveryFails(t *testing.T) {
	msg := &fakeMsg{data: paymentJSON(t)}
	c := NewJetStreamConsumer("stellar.events", "webhook-subscriber", &fakeSubs{
		subs: []models.Subscription{{ID: 1, Types: []string{"payment"}, WebhookURL: "http://x"}},
	})

	c.process(msg, func(models.Subscription, *events.Event) error {
		return errors.New("webhook down")
	})

	if msg.nakDelay != nakDelay || msg.acked {
		t.Fatalf("nak=%v acked=%v", msg.nakDelay, msg.acked)
	}
}

func TestProcessNaksIfAnyDeliveryFails(t *testing.T) {
	msg := &fakeMsg{data: paymentJSON(t)}
	c := NewJetStreamConsumer("stellar.events", "webhook-subscriber", &fakeSubs{
		subs: []models.Subscription{
			{ID: 1, Types: []string{"payment"}, WebhookURL: "http://ok"},
			{ID: 2, Types: []string{"payment"}, WebhookURL: "http://bad"},
		},
	})
	c.process(msg, func(sub models.Subscription, _ *events.Event) error {
		if sub.ID == 2 {
			return errors.New("fail")
		}
		return nil
	})

	if msg.nakDelay != nakDelay || msg.acked {
		t.Fatalf("nak=%v acked=%v", msg.nakDelay, msg.acked)
	}
}
