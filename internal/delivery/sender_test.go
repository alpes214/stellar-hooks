package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/models"
)

func testSender(serverURL string, retries int) Sender {
	return Sender{
		HTTP:           &http.Client{Timeout: time.Second},
		Sleep:          func(time.Duration) {},
		MaxRetries:     retries,
		InitialBackoff: time.Nanosecond,
	}
}

func TestSendReturnsNilOn2xx(t *testing.T) {
	var gotSig string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-Signature")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	evt := &events.Event{ID: "op-1", Type: events.EventPayment}
	err := testSender(srv.URL, 1).Send(models.Subscription{WebhookURL: srv.URL, Secret: "s"}, evt)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	body, _ := json.Marshal(evt)
	mac := hmac.New(sha256.New, []byte("s"))
	mac.Write(body)
	if gotSig != hex.EncodeToString(mac.Sum(nil)) {
		t.Fatalf("signature %q", gotSig)
	}
	if string(gotBody) != string(body) {
		t.Fatalf("body %s", gotBody)
	}
}

func TestSendReturnsErrorAfterRetries(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := testSender(srv.URL, 1).Send(
		models.Subscription{WebhookURL: srv.URL, Secret: "s"},
		&events.Event{ID: "op-1"},
	)

	if err == nil {
		t.Fatal("expected error")
	}
	if hits.Load() < 2 {
		t.Fatalf("hits %d", hits.Load())
	}
}

func TestSendTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer srv.Close()

	s := Sender{
		HTTP:           &http.Client{Timeout: 10 * time.Millisecond},
		Sleep:          func(time.Duration) {},
		MaxRetries:     0,
		InitialBackoff: time.Nanosecond,
	}
	err := s.Send(models.Subscription{WebhookURL: srv.URL, Secret: "s"}, &events.Event{ID: "op-1"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
