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

	go streamOperations(ctx, client)
	go streamPayments(ctx, client)
	go streamEffects(ctx, client)
}

func streamOperations(ctx context.Context, client *horizonclient.Client) {
	request := horizonclient.OperationRequest{Cursor: "now"}

	for {
		log.Println("Connecting to Horizon OPERATIONS stream...")

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

func streamPayments(ctx context.Context, client *horizonclient.Client) {
	request := horizonclient.OperationRequest{Cursor: "now"}

	for {
		log.Println("Connecting to Horizon PAYMENTS stream...")

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

func streamEffects(ctx context.Context, client *horizonclient.Client) {
	request := horizonclient.EffectRequest{Cursor: "now"}

	for {
		log.Println("Connecting to Horizon EFFECTS stream...")

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
