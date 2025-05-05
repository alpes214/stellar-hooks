package horizon

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/protocols/horizon/effects"
	"github.com/stellar/go/protocols/horizon/operations"

	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/stream/jetstream"
)

func StartSSEListenerJetStream() {
	cursor := "now"
	if lastCursor, err := jetstream.GetLastCursorFromStream(); err == nil && lastCursor != "" {
		cursor = lastCursor
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

	ctx := context.Background()

	go streamOperations(ctx, client, cursor)
	go streamPayments(ctx, client, cursor)
	go streamEffects(ctx, client, cursor)
}

func streamOperations(ctx context.Context, client *horizonclient.Client, cursor string) {
	request := horizonclient.OperationRequest{Cursor: cursor}

	for {
		log.Printf("Connecting to Horizon OPERATIONS stream with cursor: %s", cursor)

		producer := jetstream.NewJetStreamProducer()

		err := client.StreamOperations(ctx, request, func(op operations.Operation) {
			log.Printf("Raw Operation: %+v\n", op)

			evt, err := events.NormalizeFromHorizonOp(op)
			if err != nil {
				log.Printf("Skip unsupported op: %v", err)
				return
			}

			log.Printf("Normalized Event: %+v", evt)

			if err := producer.PublishEvent("stellar.events", evt); err != nil {
				log.Printf("Failed to publish event to JetStream: %v", err)
			}
		})

		if err != nil {
			log.Printf("Error from OPERATIONS stream: %v", err)
		}

		log.Println("Reconnecting to OPERATIONS stream in 5s...")
		time.Sleep(5 * time.Second)
	}
}

func streamPayments(ctx context.Context, client *horizonclient.Client, cursor string) {
	request := horizonclient.OperationRequest{Cursor: cursor}

	for {
		log.Printf("Connecting to Horizon PAYMENTS stream with cursor: %s", cursor)

		err := client.StreamPayments(ctx, request, func(op operations.Operation) {
			log.Printf("Raw Payment Event: %+v\n", op)
		})

		if err != nil {
			log.Printf("Error from PAYMENTS stream: %v", err)
		}

		log.Println("Reconnecting to PAYMENTS stream in 5s...")
		time.Sleep(5 * time.Second)
	}
}

func streamEffects(ctx context.Context, client *horizonclient.Client, cursor string) {
	request := horizonclient.EffectRequest{Cursor: cursor}

	for {
		log.Printf("Connecting to Horizon EFFECTS stream with cursor: %s", cursor)

		err := client.StreamEffects(ctx, request, func(effect effects.Effect) {
			log.Printf("Raw Effect Event: %+v\n", effect)
		})

		if err != nil {
			log.Printf("Error from EFFECTS stream: %v", err)
		}

		log.Println("Reconnecting to EFFECTS stream in 5s...")
		time.Sleep(5 * time.Second)
	}
}
