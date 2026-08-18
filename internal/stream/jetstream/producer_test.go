package jetstream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/alpes214/stellar-hooks/internal/events"
)

type fakeJS struct {
	msgs []*nats.Msg
	errs []error
}

func (f *fakeJS) PublishMsg(m *nats.Msg, opts ...nats.PubOpt) (*nats.PubAck, error) {
	hdr := nats.Header{}
	for k, vs := range m.Header {
		hdr[k] = append([]string(nil), vs...)
	}
	cp := nats.Msg{Subject: m.Subject, Data: append([]byte(nil), m.Data...), Header: hdr}
	f.msgs = append(f.msgs, &cp)
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		if err != nil {
			return nil, err
		}
	}
	return &nats.PubAck{Sequence: uint64(len(f.msgs))}, nil
}

func TestPublishSetsMsgId(t *testing.T) {
	js := &fakeJS{}
	p := &JetStreamProducer{pub: js, backoff: time.Millisecond}
	evt := &events.Event{ID: "op-42", Type: events.EventPayment}

	if err := p.PublishEvent(context.Background(), "stellar.events", evt); err != nil {
		t.Fatal(err)
	}

	if len(js.msgs) != 1 {
		t.Fatalf("pubs %d", len(js.msgs))
	}
	if got := js.msgs[0].Header.Get(nats.MsgIdHdr); got != "op-42" {
		t.Fatalf("msgid %q", got)
	}
	if js.msgs[0].Subject != "stellar.events" {
		t.Fatalf("subject %q", js.msgs[0].Subject)
	}
}

func TestPublishRetriesUntilSuccess(t *testing.T) {
	js := &fakeJS{errs: []error{errors.New("temp"), nil}}
	p := &JetStreamProducer{pub: js, backoff: time.Millisecond}

	if err := p.PublishEvent(context.Background(), "stellar.events", &events.Event{ID: "op-1"}); err != nil {
		t.Fatal(err)
	}
	if len(js.msgs) != 2 {
		t.Fatalf("pubs %d", len(js.msgs))
	}
}

func TestPublishStopsOnCancel(t *testing.T) {
	js := &fakeJS{errs: []error{errors.New("down"), errors.New("down")}}
	p := &JetStreamProducer{pub: js, backoff: time.Hour}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := p.PublishEvent(ctx, "stellar.events", &events.Event{ID: "op-1"})

	if err == nil {
		t.Fatal("expected error")
	}
}
