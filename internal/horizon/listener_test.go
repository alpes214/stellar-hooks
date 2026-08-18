package horizon

import (
	"context"
	"errors"
	"testing"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/base"
	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/storage"
)

type fakePublisher struct {
	err   error
	calls int
}

func (f *fakePublisher) PublishEvent(ctx context.Context, subject string, event *events.Event) error {
	f.calls++
	return f.err
}

func TestInitialCursorEmptyIsNow(t *testing.T) {
	got, err := initialCursor(storage.NewMemoryCursorStore())
	if err != nil {
		t.Fatal(err)
	}
	if got != "now" {
		t.Fatalf("got %q", got)
	}
}

func TestInitialCursorUsesStoredToken(t *testing.T) {
	s := storage.NewMemoryCursorStore()
	_ = s.SetCursor(operationsCursorStream, "pt-7")

	got, err := initialCursor(s)

	if err != nil {
		t.Fatal(err)
	}
	if got != "pt-7" {
		t.Fatalf("got %q", got)
	}
}

func TestInitialCursorPropagatesError(t *testing.T) {
	_, err := initialCursor(errCursor{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleOperationPersistsPagingToken(t *testing.T) {
	s := storage.NewMemoryCursorStore()
	req := horizonclient.OperationRequest{Cursor: "now"}
	op := operations.Payment{
		Base:   operations.Base{ID: "id-1", PT: "pt-1"},
		From:   "GSRC",
		To:     "GDST",
		Amount: "1",
		Asset:  base.Asset{Code: "XLM"},
	}

	handleOperation(context.Background(), op, &fakePublisher{}, s, &req)

	if req.Cursor != "pt-1" {
		t.Fatalf("request cursor %q", req.Cursor)
	}
	got, _ := s.GetCursor(operationsCursorStream)
	if got != "pt-1" {
		t.Fatalf("stored %q", got)
	}
}

func TestHandleOperationDoesNotAdvanceOnPublishError(t *testing.T) {
	s := storage.NewMemoryCursorStore()
	req := horizonclient.OperationRequest{Cursor: "now"}
	op := operations.Payment{
		Base:   operations.Base{ID: "id-1", PT: "pt-1"},
		From:   "GSRC",
		To:     "GDST",
		Amount: "1",
	}

	handleOperation(context.Background(), op, &fakePublisher{err: errors.New("nats")}, s, &req)

	if req.Cursor != "now" {
		t.Fatalf("request cursor %q", req.Cursor)
	}
	got, _ := s.GetCursor(operationsCursorStream)
	if got != "" {
		t.Fatalf("stored %q", got)
	}
}

func TestHandleUnsupportedOpStillAdvancesCursor(t *testing.T) {
	s := storage.NewMemoryCursorStore()
	req := horizonclient.OperationRequest{Cursor: "now"}
	op := operations.BumpSequence{Base: operations.Base{PT: "pt-skip"}}

	handleOperation(context.Background(), op, &fakePublisher{}, s, &req)

	if req.Cursor != "pt-skip" {
		t.Fatalf("request cursor %q", req.Cursor)
	}
}

type errCursor struct{}

func (errCursor) GetCursor(string) (string, error) { return "", errors.New("db") }
func (errCursor) SetCursor(string, string) error   { return nil }
