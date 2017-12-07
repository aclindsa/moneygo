package store

import (
	"github.com/aclindsa/moneygo/internal/models"
)

type SessionStore interface {
	InsertSession(session *models.Session) error
	GetSession(secret string) (*models.Session, error)
	SessionExists(secret string) (bool, error)
	DeleteSession(session *models.Session) error
}

type Tx interface {
	Commit() error
	Rollback() error

	SessionStore
}

type Store interface {
	Begin() (Tx, error)
	Close() error
}
