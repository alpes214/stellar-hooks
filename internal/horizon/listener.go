package horizon

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/effects"
	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/stream/jetstream"
)

const operationsCursorStream = "horizon-operations"

type cursorStore interface {
	GetCursor(stream string) (string, error)
	SetCursor(stream string, cursor string) error
}

type eventPublisher interface {
	PublishEvent(ctx context.Context, subject string, event *events.Event) error
}

func StartSSEListenerJetStream(ctx context.Context, store cursorStore) *sync.WaitGroup {
	cursor, err := initialCursor(store)
	if err != nil {
		log.Fatalf("Failed to load Horizon cursor: %v", err)
	}
	if cursor != "now" {
		log.Printf("Resuming Horizon stream from cursor: %s", cursor)
	}

	horizonURL := os.Getenv("HORIZON_STREAM_URL")
	if horizonURL == "" {
		horizonURL = "https://horizon.stellar.org"
	}

	client := &horizonclient.Client{
		HorizonURL: horizonURL,
		HTTP:       http.DefaultClient,
		AppName:    "stellar-hooks",
		AppVersion: "1.0.0",
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		streamOperations(ctx, client, cursor, store)
	}()
	go func() {
		defer wg.Done()
		streamPayments(ctx, client, cursor)
	}()
	go func() {
		defer wg.Done()
		streamEffects(ctx, client, cursor)
	}()
	return &wg
}

func initialCursor(store cursorStore) (string, error) {
	c, err := store.GetCursor(operationsCursorStream)
	if err != nil {
		return "", err
	}
	if c == "" {
		return "now", nil
	}
	return c, nil
}

func persistCursor(store cursorStore, request *horizonclient.OperationRequest, token string) {
	if token == "" {
		return
	}
	request.Cursor = token
	if err := store.SetCursor(operationsCursorStream, token); err != nil {
		log.Printf("Failed to persist Horizon cursor: %v", err)
	}
}

func handleOperation(ctx context.Context, op operations.Operation, pub eventPublisher, store cursorStore, request *horizonclient.OperationRequest) {
	token := op.PagingToken()
	evt, err := events.NormalizeFromHorizonOp(op)
	if err != nil {
		log.Printf("Skip unsupported op: %v", err)
		persistCursor(store, request, token)
		return
	}

	log.Printf("Normalized Event: %+v", evt)

	if err := pub.PublishEvent(ctx, "stellar.events", evt); err != nil {
		log.Printf("Failed to publish event to JetStream: %v", err)
		return
	}
	persistCursor(store, request, token)
}

func streamOperations(ctx context.Context, client *horizonclient.Client, cursor string, store cursorStore) {
	request := horizonclient.OperationRequest{Cursor: cursor}
	producer := jetstream.NewJetStreamProducer()

	for {
		if ctx.Err() != nil {
			return
		}

		log.Printf("Connecting to Horizon OPERATIONS stream with cursor: %s", request.Cursor)

		err := client.StreamOperations(ctx, request, func(op operations.Operation) {
			log.Printf("Raw Operation: %+v\n", op)
			handleOperation(ctx, op, producer, store, &request)
		})

		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("Error from OPERATIONS stream: %v", err)
		}

		log.Println("Reconnecting to OPERATIONS stream in 5s...")
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func streamPayments(ctx context.Context, client *horizonclient.Client, cursor string) {
	request := horizonclient.OperationRequest{Cursor: cursor}

	for {
		if ctx.Err() != nil {
			return
		}

		log.Printf("Connecting to Horizon PAYMENTS stream with cursor: %s", cursor)

		err := client.StreamPayments(ctx, request, func(op operations.Operation) {
			log.Printf("Raw Payment Event: %+v\n", op)
		})

		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("Error from PAYMENTS stream: %v", err)
		}

		log.Println("Reconnecting to PAYMENTS stream in 5s...")
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func streamEffects(ctx context.Context, client *horizonclient.Client, cursor string) {
	request := horizonclient.EffectRequest{Cursor: cursor}

	for {
		if ctx.Err() != nil {
			return
		}

		log.Printf("Connecting to Horizon EFFECTS stream with cursor: %s", cursor)

		err := client.StreamEffects(ctx, request, func(effect effects.Effect) {
			log.Printf("Raw Effect Event: %+v\n", effect)
		})

		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("Error from EFFECTS stream: %v", err)
		}

		log.Println("Reconnecting to EFFECTS stream in 5s...")
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
