package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/alpes214/stellar-hooks/internal/models"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Create(sub models.Subscription) (int64, error) {
	typesJSON, _ := json.Marshal(sub.Types)
	srcJSON, _ := json.Marshal(sub.SourceAccounts)
	dstJSON, _ := json.Marshal(sub.DestAccounts)

	query := `
		INSERT INTO subscriptions (
			account_id, webhook_url, secret,
			types, source_accounts, dest_accounts,
			asset_code, asset_issuer
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id;
	`
	var id int64
	err := s.db.QueryRow(
		query,
		sub.AccountID,
		sub.WebhookURL,
		sub.Secret,
		typesJSON,
		srcJSON,
		dstJSON,
		sub.AssetCode,
		sub.AssetIssuer,
	).Scan(&id)

	return id, err
}

func (s *PostgresStore) List() ([]models.Subscription, error) {
	rows, err := s.db.Query(`SELECT id, account_id, webhook_url, secret, types, source_accounts, dest_accounts, asset_code, asset_issuer FROM subscriptions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		var typesJSON, srcJSON, dstJSON []byte

		err := rows.Scan(
			&sub.ID,
			&sub.AccountID,
			&sub.WebhookURL,
			&sub.Secret,
			&typesJSON,
			&srcJSON,
			&dstJSON,
			&sub.AssetCode,
			&sub.AssetIssuer,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(typesJSON, &sub.Types)
		_ = json.Unmarshal(srcJSON, &sub.SourceAccounts)
		_ = json.Unmarshal(dstJSON, &sub.DestAccounts)

		subs = append(subs, sub)
	}

	return subs, nil
}

func (s *PostgresStore) GetByID(id int64) (*models.Subscription, error) {
	var sub models.Subscription
	var typesJSON, srcJSON, dstJSON []byte

	query := `SELECT id, account_id, webhook_url, secret, types, source_accounts, dest_accounts, asset_code, asset_issuer FROM subscriptions WHERE id = $1`
	err := s.db.QueryRow(query, id).Scan(
		&sub.ID,
		&sub.AccountID,
		&sub.WebhookURL,
		&sub.Secret,
		&typesJSON,
		&srcJSON,
		&dstJSON,
		&sub.AssetCode,
		&sub.AssetIssuer,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(typesJSON, &sub.Types)
	_ = json.Unmarshal(srcJSON, &sub.SourceAccounts)
	_ = json.Unmarshal(dstJSON, &sub.DestAccounts)

	return &sub, nil
}

func (s *PostgresStore) Update(sub models.Subscription) error {
	typesJSON, _ := json.Marshal(sub.Types)
	srcJSON, _ := json.Marshal(sub.SourceAccounts)
	dstJSON, _ := json.Marshal(sub.DestAccounts)

	query := `
		UPDATE subscriptions SET
			account_id = $1,
			webhook_url = $2,
			secret = $3,
			types = $4,
			source_accounts = $5,
			dest_accounts = $6,
			asset_code = $7,
			asset_issuer = $8
		WHERE id = $9
	`
	_, err := s.db.Exec(
		query,
		sub.AccountID,
		sub.WebhookURL,
		sub.Secret,
		typesJSON,
		srcJSON,
		dstJSON,
		sub.AssetCode,
		sub.AssetIssuer,
		sub.ID,
	)
	return err
}

func (s *PostgresStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM subscriptions WHERE id = $1`, id)
	return err
}

func (s *PostgresStore) Count() (int64, error) {
	var count int64
	err := s.db.QueryRow(`SELECT COUNT(*) FROM subscriptions`).Scan(&count)
	return count, err
}

func (s *PostgresStore) GetAllWebhookTargets() ([]models.Subscription, error) {
	return s.List()
}

func (s *PostgresStore) GetAllSubscriptions() ([]models.Subscription, error) {
	return s.List()
}

// DB Initialization
func InitPostgres() *sql.DB {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	return db
}

// DB Migration
func MigratePostgres(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id SERIAL PRIMARY KEY,
		account_id TEXT,
		webhook_url TEXT NOT NULL,
		secret TEXT NOT NULL,
		types JSONB,
		source_accounts JSONB,
		dest_accounts JSONB,
		asset_code TEXT,
		asset_issuer TEXT
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}
