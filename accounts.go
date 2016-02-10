package main

import (
	"encoding/json"
	"errors"
	"gopkg.in/gorp.v1"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	Bank       int64 = 1
	Cash             = 2
	Asset            = 3
	Liability        = 4
	Investment       = 5
	Income           = 6
	Expense          = 7
	Trading          = 8
)

type Account struct {
	AccountId         int64
	ExternalAccountId string
	UserId            int64
	SecurityId        int64
	ParentAccountId   int64 // -1 if this account is at the root
	Type              int64
	Name              string

	// monotonically-increasing account transaction version number. Used for
	// allowing a client to ensure they have a consistent version when paging
	// through transactions.
	AccountVersion int64 `json:"Version"`
}

type AccountList struct {
	Accounts *[]Account `json:"accounts"`
}

var accountTransactionsRE *regexp.Regexp
var accountImportRE *regexp.Regexp

func init() {
	accountTransactionsRE = regexp.MustCompile(`^/account/[0-9]+/transactions/?$`)
	accountImportRE = regexp.MustCompile(`^/account/[0-9]+/import/?$`)
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

func GetAccount(accountid int64, userid int64) (*Account, error) {
	var a Account

	err := DB.SelectOne(&a, "SELECT * from accounts where UserId=? AND AccountId=?", userid, accountid)
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

func GetAccounts(userid int64) (*[]Account, error) {
	var accounts []Account

	_, err := DB.Select(&accounts, "SELECT * from accounts where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// trading account for the supplied security/currency
func GetTradingAccount(userid int64, securityid int64) (*Account, error) {
	var tradingAccounts []Account //top-level 'Trading' account(s)
	var tradingAccount Account
	var accounts []Account //second-level security-specific trading account(s)
	var account Account

	transaction, err := DB.Begin()
	if err != nil {
		return nil, err
	}

	// Try to find the top-level trading account
	_, err = transaction.Select(&tradingAccounts, "SELECT * from accounts where UserId=? AND Name='Trading' AND ParentAccountId=-1 AND Type=? ORDER BY AccountId ASC LIMIT 1", userid, Trading)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	if len(tradingAccounts) == 1 {
		tradingAccount = tradingAccounts[0]
	} else {
		tradingAccount.UserId = userid
		tradingAccount.Name = "Trading"
		tradingAccount.ParentAccountId = -1
		tradingAccount.SecurityId = 840 /*USD*/ //FIXME SecurityId shouldn't matter for top-level trading account, but maybe we should grab the user's default
		tradingAccount.Type = Trading

		err = transaction.Insert(&tradingAccount)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
	}

	// Now, try to find the security-specific trading account
	_, err = transaction.Select(&accounts, "SELECT * from accounts where UserId=? AND SecurityId=? AND ParentAccountId=? ORDER BY AccountId ASC LIMIT 1", userid, securityid, tradingAccount.AccountId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	if len(accounts) == 1 {
		account = accounts[0]
	} else {
		security := GetSecurity(securityid)
		account.UserId = userid
		account.Name = security.Name
		account.ParentAccountId = tradingAccount.AccountId
		account.SecurityId = securityid
		account.Type = Trading

		err = transaction.Insert(&account)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &account, nil
}

type ParentAccountMissingError struct{}

func (pame ParentAccountMissingError) Error() string {
	return "Parent account missing"
}

func insertUpdateAccount(a *Account, insert bool) error {
	transaction, err := DB.Begin()
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

func InsertAccount(a *Account) error {
	return insertUpdateAccount(a, true)
}

func UpdateAccount(a *Account) error {
	return insertUpdateAccount(a, false)
}

func DeleteAccount(a *Account) error {
	transaction, err := DB.Begin()
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

func AccountHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		WriteError(w, 1 /*Not Signed In*/)
		return
	}

	if r.Method == "POST" {
		// if URL looks like /account/[0-9]+/import, use the account
		// import handler
		if accountImportRE.MatchString(r.URL.Path) {
			var accountid int64
			n, err := GetURLPieces(r.URL.Path, "/account/%d", &accountid)

			if err != nil || n != 1 {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
			AccountImportHandler(w, r, user, accountid)
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

		if GetSecurity(account.SecurityId) == nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		err = InsertAccount(&account)
		if err != nil {
			if _, ok := err.(ParentAccountMissingError); ok {
				WriteError(w, 3 /*Invalid Request*/)
			} else {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
			}
			return
		}

		WriteSuccess(w)
	} else if r.Method == "GET" {
		var accountid int64
		n, err := GetURLPieces(r.URL.Path, "/account/%d", &accountid)

		if err != nil || n != 1 {
			//Return all Accounts
			var al AccountList
			accounts, err := GetAccounts(user.UserId)
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
				AccountTransactionsHandler(w, r, user, accountid)
				return
			}

			// Return Account with this Id
			account, err := GetAccount(accountid, user.UserId)
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

			if GetSecurity(account.SecurityId) == nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = UpdateAccount(&account)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		} else if r.Method == "DELETE" {
			accountid, err := GetURLID(r.URL.Path)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			account, err := GetAccount(accountid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = DeleteAccount(account)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}
