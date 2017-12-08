package store

import (
	"github.com/aclindsa/moneygo/internal/models"
)

type SessionStore interface {
	SessionExists(secret string) (bool, error)
	InsertSession(session *models.Session) error
	GetSession(secret string) (*models.Session, error)
	DeleteSession(session *models.Session) error
}

type UserStore interface {
	UsernameExists(username string) (bool, error)
	InsertUser(user *models.User) error
	GetUser(userid int64) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(user *models.User) error
}

type SecurityStore interface {
	InsertSecurity(security *models.Security) error
	GetSecurity(securityid int64, userid int64) (*models.Security, error)
	GetSecurities(userid int64) (*[]*models.Security, error)
	FindMatchingSecurities(userid int64, security *models.Security) (*[]*models.Security, error)
	UpdateSecurity(security *models.Security) error
	DeleteSecurity(security *models.Security) error
}

type Tx interface {
	Commit() error
	Rollback() error

	SessionStore
	UserStore
	SecurityStore
}

type Store interface {
	Begin() (Tx, error)
	Close() error
}
