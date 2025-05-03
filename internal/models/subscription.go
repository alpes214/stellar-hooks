package models

type Subscription struct {
	ID             int64    `json:"id"`
	Types          []string `json:"types"`
	SourceAccounts []string `json:"source_accounts"`
	DestAccounts   []string `json:"dest_accounts"`
	AssetCode      string   `json:"asset_code"`
	AssetIssuer    string   `json:"asset_issuer"`
	AccountID      string   `json:"account_id"`
	WebhookURL     string   `json:"webhook_url"`
	Secret         string   `json:"secret"`
}
