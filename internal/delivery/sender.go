package delivery

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/models"
)

const (
	maxRetries     = 5
	initialBackoff = 2 * time.Second
	backoffFactor  = 2.0
	httpTimeout    = 10 * time.Second
)

type Sender struct {
	HTTP           *http.Client
	Sleep          func(time.Duration)
	MaxRetries     int
	InitialBackoff time.Duration
}

var defaultSender = Sender{
	HTTP:           &http.Client{Timeout: httpTimeout},
	Sleep:          time.Sleep,
	MaxRetries:     maxRetries,
	InitialBackoff: initialBackoff,
}

func SendToWebhook(sub models.Subscription, evt *events.Event) error {
	return defaultSender.Send(sub, evt)
}

func (s Sender) Send(sub models.Subscription, evt *events.Event) error {
	body, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(sub.Secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	client := s.HTTP
	if client == nil {
		client = defaultSender.HTTP
	}
	sleep := s.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	retries := s.MaxRetries
	backoff := s.InitialBackoff
	if backoff == 0 {
		backoff = initialBackoff
	}

	var attempt int
	for {
		req, err := http.NewRequest("POST", sub.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Signature", signature)

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Webhook sent to %s with status %s", sub.WebhookURL, resp.Status)
			resp.Body.Close()
			return nil
		}

		if resp != nil {
			if err == nil {
				err = fmt.Errorf("status %s", resp.Status)
			}
			resp.Body.Close()
		}

		attempt++
		if attempt > retries {
			return fmt.Errorf("webhook delivery failed after %d retries: %w", attempt, err)
		}

		log.Printf("Retry %d for webhook to %s after %v", attempt, sub.WebhookURL, backoff)
		sleep(backoff)
		backoff = time.Duration(float64(backoff) * backoffFactor)
	}
}
