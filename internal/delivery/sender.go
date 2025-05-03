package delivery

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
)

func SendToWebhook(sub models.Subscription, evt *events.Event) {
	body, err := json.Marshal(evt)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	mac := hmac.New(sha256.New, []byte(sub.Secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	var retries int
	var backoff = initialBackoff

	for {
		req, err := http.NewRequest("POST", sub.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Failed to create webhook request: %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Signature", signature)

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("Webhook sent to %s with status %s", sub.WebhookURL, resp.Status)
			resp.Body.Close()
			return
		}

		// If error or non-2xx status
		if resp != nil {
			resp.Body.Close()
		}

		retries++
		if retries > maxRetries {
			log.Printf("Webhook delivery failed after %d retries: %v", retries, err)
			return
		}

		log.Printf("Retry %d for webhook to %s after %v", retries, sub.WebhookURL, backoff)
		time.Sleep(backoff)
		backoff = time.Duration(float64(backoff) * backoffFactor) // Exponential increase
	}
}
