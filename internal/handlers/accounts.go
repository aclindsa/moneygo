package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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

func (al *AccountList) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(al)
}

func GetAccount(tx *Tx, accountid int64, userid int64) (*Account, error) {
	var a Account

	err := tx.SelectOne(&a, "SELECT * from accounts where UserId=? AND AccountId=?", userid, accountid)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func GetAccounts(tx *Tx, userid int64) (*[]Account, error) {
	var accounts []Account

	_, err := tx.Select(&accounts, "SELECT * from accounts where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

// Get (and attempt to create if it doesn't exist). Matches on UserId,
// SecurityId, Type, Name, and ParentAccountId
func GetCreateAccount(tx *Tx, a Account) (*Account, error) {
	var accounts []Account
	var account Account

	// Try to find the top-level trading account
	_, err := tx.Select(&accounts, "SELECT * from accounts where UserId=? AND SecurityId=? AND Type=? AND Name=? AND ParentAccountId=? ORDER BY AccountId ASC LIMIT 1", a.UserId, a.SecurityId, a.Type, a.Name, a.ParentAccountId)
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

		err = tx.Insert(&account)
		if err != nil {
			return nil, err
		}
	}
	return &account, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// trading account for the supplied security/currency
func GetTradingAccount(tx *Tx, userid int64, securityid int64) (*Account, error) {
	var tradingAccount Account
	var account Account

	user, err := GetUser(tx, userid)
	if err != nil {
		return nil, err
	}

	tradingAccount.UserId = userid
	tradingAccount.Type = Trading
	tradingAccount.Name = "Trading"
	tradingAccount.SecurityId = user.DefaultCurrency
	tradingAccount.ParentAccountId = -1

	// Find/create the top-level trading account
	ta, err := GetCreateAccount(tx, tradingAccount)
	if err != nil {
		return nil, err
	}

	security, err := GetSecurity(tx, securityid, userid)
	if err != nil {
		return nil, err
	}

	account.UserId = userid
	account.Name = security.Name
	account.ParentAccountId = ta.AccountId
	account.SecurityId = securityid
	account.Type = Trading

	a, err := GetCreateAccount(tx, account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Get (and attempt to create if it doesn't exist) the security/currency
// imbalance account for the supplied security/currency
func GetImbalanceAccount(tx *Tx, userid int64, securityid int64) (*Account, error) {
	var imbalanceAccount Account
	var account Account
	xxxtemplate := FindSecurityTemplate("XXX", Currency)
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
	imbalanceAccount.Type = Bank

	// Find/create the top-level trading account
	ia, err := GetCreateAccount(tx, imbalanceAccount)
	if err != nil {
		return nil, err
	}

	security, err := GetSecurity(tx, securityid, userid)
	if err != nil {
		return nil, err
	}

	account.UserId = userid
	account.Name = security.Name
	account.ParentAccountId = ia.AccountId
	account.SecurityId = securityid
	account.Type = Bank

	a, err := GetCreateAccount(tx, account)
	if err != nil {
		return nil, err
	}

	return a, nil
}

type ParentAccountMissingError struct{}

func (pame ParentAccountMissingError) Error() string {
	return "Parent account missing"
}

type TooMuchNestingError struct{}

func (tmne TooMuchNestingError) Error() string {
	return "Too much nesting"
}

type CircularAccountsError struct{}

func (cae CircularAccountsError) Error() string {
	return "Would result in circular account relationship"
}

func insertUpdateAccount(tx *Tx, a *Account, insert bool) error {
	found := make(map[int64]bool)
	if !insert {
		found[a.AccountId] = true
	}
	parentid := a.ParentAccountId
	depth := 0
	for parentid != -1 {
		depth += 1
		if depth > 100 {
			return TooMuchNestingError{}
		}

		var a Account
		err := tx.SelectOne(&a, "SELECT * from accounts where AccountId=?", parentid)
		if err != nil {
			return ParentAccountMissingError{}
		}

		// Insertion by itself can never result in circular dependencies
		if insert {
			break
		}

		found[parentid] = true
		parentid = a.ParentAccountId
		if _, ok := found[parentid]; ok {
			return CircularAccountsError{}
		}
	}

	if insert {
		err := tx.Insert(a)
		if err != nil {
			return err
		}
	} else {
		oldacct, err := GetAccount(tx, a.AccountId, a.UserId)
		if err != nil {
			return err
		}

		a.AccountVersion = oldacct.AccountVersion + 1

		count, err := tx.Update(a)
		if err != nil {
			return err
		}
		if count != 1 {
			return errors.New("Updated more than one account")
		}
	}

	return nil
}

func InsertAccount(tx *Tx, a *Account) error {
	return insertUpdateAccount(tx, a, true)
}

func UpdateAccount(tx *Tx, a *Account) error {
	return insertUpdateAccount(tx, a, false)
}

func DeleteAccount(tx *Tx, a *Account) error {
	if a.ParentAccountId != -1 {
		// Re-parent splits to this account's parent account if this account isn't a root account
		_, err := tx.Exec("UPDATE splits SET AccountId=? WHERE AccountId=?", a.ParentAccountId, a.AccountId)
		if err != nil {
			return err
		}
	} else {
		// Delete splits if this account is a root account
		_, err := tx.Exec("DELETE FROM splits WHERE AccountId=?", a.AccountId)
		if err != nil {
			return err
		}
	}

	// Re-parent child accounts to this account's parent account
	_, err := tx.Exec("UPDATE accounts SET ParentAccountId=? WHERE ParentAccountId=?", a.ParentAccountId, a.AccountId)
	if err != nil {
		return err
	}

	count, err := tx.Delete(a)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Was going to delete more than one account")
	}

	return nil
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

		account_json := r.PostFormValue("account")
		if account_json == "" {
			return NewError(3 /*Invalid Request*/)
		}

		var account Account
		err := account.Read(account_json)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		account.AccountId = -1
		account.UserId = user.UserId
		account.AccountVersion = 0

		security, err := GetSecurity(context.Tx, account.SecurityId, user.UserId)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}
		if security == nil {
			return NewError(3 /*Invalid Request*/)
		}

		err = InsertAccount(context.Tx, &account)
		if err != nil {
			if _, ok := err.(ParentAccountMissingError); ok {
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
			var al AccountList
			accounts, err := GetAccounts(context.Tx, user.UserId)
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
			account, err := GetAccount(context.Tx, accountid, user.UserId)
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
			account_json := r.PostFormValue("account")
			if account_json == "" {
				return NewError(3 /*Invalid Request*/)
			}

			var account Account
			err := account.Read(account_json)
			if err != nil || account.AccountId != accountid {
				return NewError(3 /*Invalid Request*/)
			}
			account.UserId = user.UserId

			security, err := GetSecurity(context.Tx, account.SecurityId, user.UserId)
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

			err = UpdateAccount(context.Tx, &account)
			if err != nil {
				if _, ok := err.(ParentAccountMissingError); ok {
					return NewError(3 /*Invalid Request*/)
				} else if _, ok := err.(CircularAccountsError); ok {
					return NewError(3 /*Invalid Request*/)
				} else {
					log.Print(err)
					return NewError(999 /*Internal Error*/)
				}
			}

			return &account
		} else if r.Method == "DELETE" {
			account, err := GetAccount(context.Tx, accountid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = DeleteAccount(context.Tx, account)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
