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

type SecurityInUseError struct {
	Message string
}

func (e SecurityInUseError) Error() string {
	return e.Message
}

type SecurityStore interface {
	InsertSecurity(security *models.Security) error
	GetSecurity(securityid int64, userid int64) (*models.Security, error)
	GetSecurities(userid int64) (*[]*models.Security, error)
	FindMatchingSecurities(security *models.Security) (*[]*models.Security, error)
	UpdateSecurity(security *models.Security) error
	DeleteSecurity(security *models.Security) error
}

type ParentAccountMissingError struct{}

func (pame ParentAccountMissingError) Error() string {
	return "Parent account missing"
}

type TooMuchNestingError struct{}

func (tmne TooMuchNestingError) Error() string {
	return "Too much account nesting"
}

type CircularAccountsError struct{}

func (cae CircularAccountsError) Error() string {
	return "Would result in circular account relationship"
}

type AccountStore interface {
	InsertAccount(account *models.Account) error
	GetAccount(accountid int64, userid int64) (*models.Account, error)
	GetAccounts(userid int64) (*[]*models.Account, error)
	FindMatchingAccounts(account *models.Account) (*[]*models.Account, error)
	UpdateAccount(account *models.Account) error
	DeleteAccount(account *models.Account) error
}

type Tx interface {
	Commit() error
	Rollback() error

	SessionStore
	UserStore
	SecurityStore
	AccountStore
}

type Store interface {
	Begin() (Tx, error)
	Close() error
}
