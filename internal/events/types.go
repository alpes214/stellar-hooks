package events

import "time"

type EventType string

const (
	EventPayment                 EventType = "payment"
	EventAccountCreated          EventType = "account_created"
	EventTrustlineCreated        EventType = "trustline_created"
	EventTrustlineRemoved        EventType = "trustline_removed"
	EventClaimableBalanceCreated EventType = "claimable_balance_created"
	EventClaimableBalanceClaimed EventType = "claimable_balance_claimed"
)

type Asset struct {
	Code   string
	Issuer string
}

type Event struct {
	ID            string
	Type          EventType
	Ledger        int64
	Timestamp     time.Time
	SourceAccount string
	Destination   string
	Asset         *Asset
	Amount        string
	Memo          *string
	TransactionID string
	Raw           any
}
