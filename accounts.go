package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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
)

type Account struct {
	AccountId       int64
	UserId          int64
	SecurityId      int64
	ParentAccountId int64 // -1 if this account is at the root
	Type            int64
	Name            string
}

type AccountList struct {
	Accounts *[]Account `json:"accounts"`
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

func GetAccounts(userid int64) (*[]Account, error) {
	var accounts []Account

	_, err := DB.Select(&accounts, "SELECT * from accounts where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
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
		accountid, err := GetURLID(r.URL.Path)
		if err != nil {
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
