package handlers

import (
	"encoding/json"
	"errors"
	"gopkg.in/gorp.v1"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type AccountType int64

const (
	Bank       AccountType = 1 // start at 1 so that the default (0) is invalid
	Cash                   = 2
	Asset                  = 3
	Liability              = 4
	Investment             = 5
	Income                 = 6
	Expense                = 7
	Trading                = 8
	Equity                 = 9
	Receivable             = 10
	Payable                = 11
)

var AccountTypes = []AccountType{
	Bank,
	Cash,
	Asset,
	Liability,
	Investment,
	Income,
	Expense,
	Trading,
	Equity,
	Receivable,
	Payable,
}

func (t AccountType) String() string {
	switch t {
	case Bank:
		return "Bank"
	case Cash:
		return "Cash"
	case Asset:
		return "Asset"
	case Liability:
		return "Liability"
	case Investment:
		return "Investment"
	case Income:
		return "Income"
	case Expense:
		return "Expense"
	case Trading:
		return "Trading"
	case Equity:
		return "Equity"
	case Receivable:
		return "Receivable"
	case Payable:
		return "Payable"
	}
	return ""
}

type Account struct {
	AccountId         int64
	ExternalAccountId string
	UserId            int64
	SecurityId        int64
	ParentAccountId   int64 // -1 if this account is at the root
	Type              AccountType
	Name              string

	// monotonically-increasing account transaction version number. Used for
	// allowing a client to ensure they have a consistent version when paging
	// through transactions.
	AccountVersion int64 `json:"Version"`

	// Optional fields specifying how to fetch transactions from a bank via OFX
	OFXURL       string
	OFXORG       string
	OFXFID       string
	OFXUser      string
	OFXBankID    string // OFX BankID (BrokerID if AcctType == Investment)
	OFXAcctID    string
	OFXAcctType  string // ofxgo.acctType
	OFXClientUID string
	OFXAppID     string
	OFXAppVer    string
	OFXVersion   string
	OFXNoIndent  bool
}

type AccountList struct {
	Accounts *[]Account `json:"accounts"`
}

var accountTransactionsRE *regexp.Regexp
var accountImportRE *regexp.Regexp

func init() {
	accountTransactionsRE = regexp.MustCompile(`^/account/[0-9]+/transactions/?$`)
	accountImportRE = regexp.MustCompile(`^/account/[0-9]+/import/[a-z]+/?$`)
}

func (a *Account) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(a)
}

func (a *Account) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(a)
}

func (al *AccountList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(al)
}

func GetAccount(db *DB, accountid int64, userid int64) (*Account, error) {
	var a Account

	err := db.SelectOne(&a, "SELECT * from accounts where UserId=? AND AccountId=?", userid, accountid)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func GetAccountTx(transaction *gorp.Transaction, accountid int64, userid int64) (*Account, error) {
	var a Account

	err := transaction.SelectOne(&a, "SELECT * from accounts where UserId=? AND AccountId=?", userid, accountid)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

func GetAccounts(db *DB, userid int64) (*[]Account, error) {
	var accounts []Account

	_, err := db.Select(&accounts, "SELECT * from accounts where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

// Get (and attempt to create if it doesn't exist). Matches on UserId,
// SecurityId, Type, Name, and ParentAccountId
func GetCreateAccountTx(transaction *gorp.Transaction, a Account) (*Account, error) {
	var accounts []Account
	var account Account

	// Try to find the top-level trading account
	_, err := transaction.Select(&accounts, "SELECT * from accounts where UserId=? AND SecurityId=? AND Type=? AND Name=? AND ParentAccountId=? ORDER BY AccountId ASC LIMIT 1", a.UserId, a.SecurityId, a.Type, a.Name, a.ParentAccountId)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 1 {
		account = accounts[0]
	} else {
		account.UserId = a.UserId
		account.SecurityId = a.SecurityId
		account.Type = a.Type
		account.Name = a.Name
		account.ParentAccountId = a.ParentAccountId

		err = transaction.Insert(&account)
		if err != nil {
			return nil, err
		}
	}
	return &account, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// trading account for the supplied security/currency
func GetTradingAccount(transaction *gorp.Transaction, userid int64, securityid int64) (*Account, error) {
	var tradingAccount Account
	var account Account

	user, err := GetUserTx(transaction, userid)
	if err != nil {
		return nil, err
	}

	tradingAccount.UserId = userid
	tradingAccount.Type = Trading
	tradingAccount.Name = "Trading"
	tradingAccount.SecurityId = user.DefaultCurrency
	tradingAccount.ParentAccountId = -1

	// Find/create the top-level trading account
	ta, err := GetCreateAccountTx(transaction, tradingAccount)
	if err != nil {
		return nil, err
	}

	security, err := GetSecurityTx(transaction, securityid, userid)
	if err != nil {
		return nil, err
	}

	account.UserId = userid
	account.Name = security.Name
	account.ParentAccountId = ta.AccountId
	account.SecurityId = securityid
	account.Type = Trading

	a, err := GetCreateAccountTx(transaction, account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// imbalance account for the supplied security/currency
func GetImbalanceAccount(transaction *gorp.Transaction, userid int64, securityid int64) (*Account, error) {
	var imbalanceAccount Account
	var account Account
	xxxtemplate := FindSecurityTemplate("XXX", Currency)
	if xxxtemplate == nil {
		return nil, errors.New("Couldn't find XXX security template")
	}
	xxxsecurity, err := ImportGetCreateSecurity(transaction, userid, xxxtemplate)
	if err != nil {
		return nil, errors.New("Couldn't create XXX security")
	}

	imbalanceAccount.UserId = userid
	imbalanceAccount.Name = "Imbalances"
	imbalanceAccount.ParentAccountId = -1
	imbalanceAccount.SecurityId = xxxsecurity.SecurityId
	imbalanceAccount.Type = Bank

	// Find/create the top-level trading account
	ia, err := GetCreateAccountTx(transaction, imbalanceAccount)
	if err != nil {
		return nil, err
	}

	security, err := GetSecurityTx(transaction, securityid, userid)
	if err != nil {
		return nil, err
	}

	account.UserId = userid
	account.Name = security.Name
	account.ParentAccountId = ia.AccountId
	account.SecurityId = securityid
	account.Type = Bank

	a, err := GetCreateAccountTx(transaction, account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

type ParentAccountMissingError struct{}

func (pame ParentAccountMissingError) Error() string {
	return "Parent account missing"
}

func insertUpdateAccount(db *DB, a *Account, insert bool) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	if a.ParentAccountId != -1 {
		existing, err := transaction.SelectInt("SELECT count(*) from accounts where AccountId=?", a.ParentAccountId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if existing != 1 {
			transaction.Rollback()
			return ParentAccountMissingError{}
		}
	}

	if insert {
		err = transaction.Insert(a)
		if err != nil {
			transaction.Rollback()
			return err
		}
	} else {
		oldacct, err := GetAccountTx(transaction, a.AccountId, a.UserId)
		if err != nil {
			transaction.Rollback()
			return err
		}

		a.AccountVersion = oldacct.AccountVersion + 1

		count, err := transaction.Update(a)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if count != 1 {
			transaction.Rollback()
			return errors.New("Updated more than one account")
		}
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func InsertAccount(db *DB, a *Account) error {
	return insertUpdateAccount(db, a, true)
}

func UpdateAccount(db *DB, a *Account) error {
	return insertUpdateAccount(db, a, false)
}

func DeleteAccount(db *DB, a *Account) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	if a.ParentAccountId != -1 {
		// Re-parent splits to this account's parent account if this account isn't a root account
		_, err = transaction.Exec("UPDATE splits SET AccountId=? WHERE AccountId=?", a.ParentAccountId, a.AccountId)
		if err != nil {
			transaction.Rollback()
			return err
		}
	} else {
		// Delete splits if this account is a root account
		_, err = transaction.Exec("DELETE FROM splits WHERE AccountId=?", a.AccountId)
		if err != nil {
			transaction.Rollback()
			return err
		}
	}

	// Re-parent child accounts to this account's parent account
	_, err = transaction.Exec("UPDATE accounts SET ParentAccountId=? WHERE ParentAccountId=?", a.ParentAccountId, a.AccountId)
	if err != nil {
		transaction.Rollback()
		return err
	}

	count, err := transaction.Delete(a)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if count != 1 {
		transaction.Rollback()
		return errors.New("Was going to delete more than one account")
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func AccountHandler(w http.ResponseWriter, r *http.Request, db *DB) {
	user, err := GetUserFromSession(db, r)
	if err != nil {
		WriteError(w, 1 /*Not Signed In*/)
		return
	}

	if r.Method == "POST" {
		// if URL looks like /account/[0-9]+/import, use the account
		// import handler
		if accountImportRE.MatchString(r.URL.Path) {
			var accountid int64
			var importtype string
			n, err := GetURLPieces(r.URL.Path, "/account/%d/import/%s", &accountid, &importtype)

			if err != nil || n != 2 {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
			AccountImportHandler(db, w, r, user, accountid, importtype)
			return
		}

		account_json := r.PostFormValue("account")
		if account_json == "" {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		var account Account
		err := account.Read(account_json)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		account.AccountId = -1
		account.UserId = user.UserId
		account.AccountVersion = 0

		security, err := GetSecurity(db, account.SecurityId, user.UserId)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
		if security == nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		err = InsertAccount(db, &account)
		if err != nil {
			if _, ok := err.(ParentAccountMissingError); ok {
				WriteError(w, 3 /*Invalid Request*/)
			} else {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
			}
			return
		}

		w.WriteHeader(201 /*Created*/)
		err = account.Write(w)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	} else if r.Method == "GET" {
		var accountid int64
		n, err := GetURLPieces(r.URL.Path, "/account/%d", &accountid)

		if err != nil || n != 1 {
			//Return all Accounts
			var al AccountList
			accounts, err := GetAccounts(db, user.UserId)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
			al.Accounts = accounts
			err = (&al).Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else {
			// if URL looks like /account/[0-9]+/transactions, use the account
			// transaction handler
			if accountTransactionsRE.MatchString(r.URL.Path) {
				AccountTransactionsHandler(db, w, r, user, accountid)
				return
			}

			// Return Account with this Id
			account, err := GetAccount(db, accountid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = account.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		}
	} else {
		accountid, err := GetURLID(r.URL.Path)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		if r.Method == "PUT" {
			account_json := r.PostFormValue("account")
			if account_json == "" {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			var account Account
			err := account.Read(account_json)
			if err != nil || account.AccountId != accountid {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			account.UserId = user.UserId

			security, err := GetSecurity(db, account.SecurityId, user.UserId)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
			if security == nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = UpdateAccount(db, &account)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			err = account.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else if r.Method == "DELETE" {
			account, err := GetAccount(db, accountid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = DeleteAccount(db, account)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}
