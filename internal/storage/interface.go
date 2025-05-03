package storage

import "github.com/alpes214/stellar-hooks/internal/models"

type SubscriptionStore interface {
	GetAllSubscriptions() ([]models.Subscription, error)
	GetAllWebhookTargets() ([]models.Subscription, error)
	List() ([]models.Subscription, error)
	Create(sub models.Subscription) (int64, error)
	GetByID(id int64) (*models.Subscription, error)
	Update(sub models.Subscription) error
	Delete(id int64) error
	Count() (int64, error)
}

type CursorStore interface {
	GetCursor(stream string) (string, error)
	SetCursor(stream string, cursor string) error
	IsProcessed(stream string, id string) (bool, error)
	MarkProcessed(stream string, id string) error
}
