package filter

import (
	"github.com/alpes214/stellar-hooks/internal/events"
	"github.com/alpes214/stellar-hooks/internal/models"
)

func Matches(sub *models.Subscription, evt *events.Event) bool {
	if !contains(sub.Types, string(evt.Type)) {
		return false
	}
	if len(sub.SourceAccounts) > 0 && !contains(sub.SourceAccounts, evt.SourceAccount) {
		return false
	}
	if len(sub.DestAccounts) > 0 && !contains(sub.DestAccounts, evt.Destination) {
		return false
	}
	if sub.AssetCode != "" && (evt.Asset == nil || evt.Asset.Code != sub.AssetCode) {
		return false
	}
	if sub.AssetIssuer != "" && (evt.Asset == nil || evt.Asset.Issuer != sub.AssetIssuer) {
		return false
	}
	// TODO: Add amount checks, memo, etc.
	return true
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
