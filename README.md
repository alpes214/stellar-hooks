# StellarHooks – Real-Time Blockchain Webhook Notifications for Stellar

StellarHooks is a lightweight, developer-focused webhook delivery platform for the Stellar blockchain.  
It transforms Stellar Horizon SSE streams into secure, real-time HTTP webhooks with automatic retry logic and a simple subscription API.

=======
## MVP Features

- Real-time Stellar event processing via Horizon SSE:
  - `payment`, `create_account`, `change_trust`, `account_merge`
- REST API to manage subscriptions:
  - `POST`, `GET`, `DELETE /subscriptions`
- Secure webhook delivery with HMAC-SHA256 signatures
- Exponential backoff retry logic for failed deliveries
- Configurable via `.env` and runs in Docker

## Tech Stack

- Go + Gin
- PostgreSQL (via Docker)
- Stellar Horizon API (publicnet or testnet)
- Docker & Docker Compose

## Quick Start

Start the service:

```bash
docker compose up --build
```

To create a subscription:

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "G...YOURACCOUNT",
    "webhook_url": "https://webhook.site/your-url",
    "secret": "supersecretkey"
  }'
```

## Project Goals

- Help developers integrate Stellar without managing SSE streams
- Provide a secure, scalable webhook infrastructure for wallets, anchors, and fintechs
- Deliver real-time events from the Stellar blockchain to client servers

---

Built for the Stellar ecosystem.  
GitHub: [https://github.com/yourusername/stellar-hooks](https://github.com/yourusername/stellar-hooks)
