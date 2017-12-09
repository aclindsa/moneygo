package handlers

import (
	"errors"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"log"
	"net/http"
)

// Get (and attempt to create if it doesn't exist). Matches on UserId,
// SecurityId, Type, Name, and ParentAccountId
func GetCreateAccount(tx store.Tx, a models.Account) (*models.Account, error) {
	var account models.Account

	accounts, err := tx.FindMatchingAccounts(&a)
	if err != nil {
		return nil, err
	}
	if len(*accounts) > 0 {
		account = *(*accounts)[0]
	} else {
		account.UserId = a.UserId
		account.SecurityId = a.SecurityId
		account.Type = a.Type
		account.Name = a.Name
		account.ParentAccountId = a.ParentAccountId

		err = tx.InsertAccount(&account)
		if err != nil {
			return nil, err
		}
	}
	return &account, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// trading account for the supplied security/currency
func GetTradingAccount(tx store.Tx, userid int64, securityid int64) (*models.Account, error) {
	var tradingAccount models.Account
	var account models.Account

	user, err := tx.GetUser(userid)
	if err != nil {
		return nil, err
	}

	tradingAccount.UserId = userid
	tradingAccount.Type = models.Trading
	tradingAccount.Name = "Trading"
	tradingAccount.SecurityId = user.DefaultCurrency
	tradingAccount.ParentAccountId = -1

	// Find/create the top-level trading account
	ta, err := GetCreateAccount(tx, tradingAccount)
	if err != nil {
		return nil, err
	}

	security, err := tx.GetSecurity(securityid, userid)
	if err != nil {
		return nil, err
	}

	account.UserId = userid
	account.Name = security.Name
	account.ParentAccountId = ta.AccountId
	account.SecurityId = securityid
	account.Type = models.Trading

	a, err := GetCreateAccount(tx, account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// imbalance account for the supplied security/currency
func GetImbalanceAccount(tx store.Tx, userid int64, securityid int64) (*models.Account, error) {
	var imbalanceAccount models.Account
	var account models.Account
	xxxtemplate := FindSecurityTemplate("XXX", models.Currency)
	if xxxtemplate == nil {
		return nil, errors.New("Couldn't find XXX security template")
	}
	xxxsecurity, err := ImportGetCreateSecurity(tx, userid, xxxtemplate)
	if err != nil {
		return nil, errors.New("Couldn't create XXX security")
	}

	imbalanceAccount.UserId = userid
	imbalanceAccount.Name = "Imbalances"
	imbalanceAccount.ParentAccountId = -1
	imbalanceAccount.SecurityId = xxxsecurity.SecurityId
	imbalanceAccount.Type = models.Bank

	// Find/create the top-level trading account
	ia, err := GetCreateAccount(tx, imbalanceAccount)
	if err != nil {
		return nil, err
	}

	security, err := tx.GetSecurity(securityid, userid)
	if err != nil {
		return nil, err
	}

	account.UserId = userid
	account.Name = security.Name
	account.ParentAccountId = ia.AccountId
	account.SecurityId = securityid
	account.Type = models.Bank

	a, err := GetCreateAccount(tx, account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func AccountHandler(r *http.Request, context *Context) ResponseWriterWriter {
	user, err := GetUserFromSession(context.Tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method == "POST" {
		if !context.LastLevel() {
			accountid, err := context.NextID()
			if err != nil || context.NextLevel() != "imports" {
				return NewError(3 /*Invalid Request*/)
			}
			return AccountImportHandler(context, r, user, accountid)
		}

		var account models.Account
		if err := ReadJSON(r, &account); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		account.AccountId = -1
		account.UserId = user.UserId
		account.AccountVersion = 0

		security, err := context.Tx.GetSecurity(account.SecurityId, user.UserId)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
		if security == nil {
			return NewError(3 /*Invalid Request*/)
		}

		err = context.Tx.InsertAccount(&account)
		if err != nil {
			if _, ok := err.(store.ParentAccountMissingError); ok {
				return NewError(3 /*Invalid Request*/)
			} else {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
		}

		return ResponseWrapper{201, &account}
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all Accounts
			var al models.AccountList
			accounts, err := context.Tx.GetAccounts(user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			al.Accounts = accounts
			return &al
		}

		accountid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		if context.LastLevel() {
			// Return Account with this Id
			account, err := context.Tx.GetAccount(accountid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			return account
		} else if context.NextLevel() == "transactions" {
			return AccountTransactionsHandler(context, r, user, accountid)
		}
	} else {
		accountid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		if r.Method == "PUT" {
			var account models.Account
			if err := ReadJSON(r, &account); err != nil || account.AccountId != accountid {
				return NewError(3 /*Invalid Request*/)
			}
			account.UserId = user.UserId

			security, err := context.Tx.GetSecurity(account.SecurityId, user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			if security == nil {
				return NewError(3 /*Invalid Request*/)
			}

			if account.ParentAccountId == account.AccountId {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.UpdateAccount(&account)
			if err != nil {
				if _, ok := err.(store.ParentAccountMissingError); ok {
					return NewError(3 /*Invalid Request*/)
				} else if _, ok := err.(store.CircularAccountsError); ok {
					return NewError(3 /*Invalid Request*/)
				} else {
					log.Print(err)
					return NewError(999 /*Internal Error*/)
				}
			}

			return &account
		} else if r.Method == "DELETE" {
			account, err := context.Tx.GetAccount(accountid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.DeleteAccount(account)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
