package store

import (
	"github.com/aclindsa/moneygo/internal/models"
	"time"
)

type UserStore interface {
	UsernameExists(username string) (bool, error)
	InsertUser(user *models.User) error
	GetUser(userid int64) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(user *models.User) error
}

type SessionStore interface {
	SessionExists(secret string) (bool, error)
	InsertSession(session *models.Session) error
	GetSession(secret string) (*models.Session, error)
	DeleteSession(session *models.Session) error
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

type PriceStore interface {
	PriceExists(price *models.Price) (bool, error)
	InsertPrice(price *models.Price) error
	GetPrice(priceid, securityid int64) (*models.Price, error)
	GetPrices(securityid int64) (*[]*models.Price, error)
	GetLatestPrice(security, currency *models.Security, date *time.Time) (*models.Price, error)
	GetEarliestPrice(security, currency *models.Security, date *time.Time) (*models.Price, error)
	UpdatePrice(price *models.Price) error
	DeletePrice(price *models.Price) error
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

type AccountMissingError struct{}

func (ame AccountMissingError) Error() string {
	return "Account missing"
}

type TransactionStore interface {
	SplitExists(s *models.Split) (bool, error)
	InsertTransaction(t *models.Transaction, user *models.User) error
	GetTransaction(transactionid int64, userid int64) (*models.Transaction, error)
	GetTransactions(userid int64) (*[]*models.Transaction, error)
	UpdateTransaction(t *models.Transaction, user *models.User) error
	DeleteTransaction(t *models.Transaction, user *models.User) error
	GetAccountSplits(user *models.User, accountid int64) (*[]*models.Split, error)
	GetAccountSplitsDate(user *models.User, accountid int64, date *time.Time) (*[]*models.Split, error)
	GetAccountSplitsDateRange(user *models.User, accountid int64, begin, end *time.Time) (*[]*models.Split, error)
	GetAccountTransactions(user *models.User, accountid int64, sort string, page uint64, limit uint64) (*models.AccountTransactionsList, error)
}

type ReportStore interface {
}

type Tx interface {
	Commit() error
	Rollback() error

	UserStore
	SessionStore
	SecurityStore
	PriceStore
	AccountStore
	TransactionStore
	ReportStore
}

type Store interface {
	Begin() (Tx, error)
	Close() error
}
