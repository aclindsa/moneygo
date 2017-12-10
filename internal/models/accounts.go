package models

import (
	"encoding/json"
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
	Accounts *[]*Account `json:"accounts"`
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
