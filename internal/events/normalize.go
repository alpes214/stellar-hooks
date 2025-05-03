package events

import (
	"fmt"

	"github.com/stellar/go/protocols/horizon/operations"
)

// NormalizeFromHorizonOp converts a Horizon operation into an internal Event struct.
func NormalizeFromHorizonOp(op operations.Operation) (*Event, error) {
	switch p := op.(type) {
	case operations.Payment:
		return &Event{
			ID:            p.ID,
			Type:          EventPayment,
			SourceAccount: p.From,
			Destination:   p.To,
			Asset: &Asset{
				Code:   p.Asset.Code,
				Issuer: p.Asset.Issuer,
			},
			Amount:        p.Amount,
			TransactionID: p.TransactionHash,
			Raw:           p,
		}, nil

	case operations.CreateAccount:
		return &Event{
			ID:            p.ID,
			Type:          EventAccountCreated,
			SourceAccount: p.Funder,
			Destination:   p.Account,
			Amount:        p.StartingBalance,
			TransactionID: p.TransactionHash,
			Raw:           p,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operation type: %T", op)
	}
}
