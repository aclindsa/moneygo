package models

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Split.Status
const (
	Imported   int64 = 1
	Entered          = 2
	Cleared          = 3
	Reconciled       = 4
	Voided           = 5
)

// Split.ImportSplitType
const (
	Default         int64 = 0
	ImportAccount         = 1 // This split belongs to the main account being imported
	SubAccount            = 2 // This split belongs to a sub-account of that being imported
	ExternalAccount       = 3
	TradingAccount        = 4
	Commission            = 5
	Taxes                 = 6
	Fees                  = 7
	Load                  = 8
	IncomeAccount         = 9
	ExpenseAccount        = 10
)

type Split struct {
	SplitId         int64
	TransactionId   int64
	Status          int64
	ImportSplitType int64

	// One of AccountId and SecurityId must be -1
	// In normal splits, AccountId will be valid and SecurityId will be -1. The
	// only case where this is reversed is for transactions that have been
	// imported and not yet associated with an account.
	AccountId  int64
	SecurityId int64

	RemoteId string // unique ID from server, for detecting duplicates
	Number   string // Check or reference number
	Memo     string
	Amount   Amount
}

func (s *Split) Valid() bool {
	return (s.AccountId == -1) != (s.SecurityId == -1)
}

type Transaction struct {
	TransactionId int64
	UserId        int64
	Description   string
	Date          time.Time
	Splits        []*Split `db:"-"`
}

type TransactionList struct {
	Transactions *[]*Transaction `json:"transactions"`
}

type AccountTransactionsList struct {
	Account           *Account
	Transactions      *[]*Transaction
	TotalTransactions int64
	BeginningBalance  Amount
	EndingBalance     Amount
}

func (t *Transaction) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(t)
}

func (t *Transaction) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(t)
}

func (tl *TransactionList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(tl)
}

func (tl *TransactionList) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(tl)
}

func (atl *AccountTransactionsList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(atl)
}

func (atl *AccountTransactionsList) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(atl)
}

func (t *Transaction) Valid() bool {
	for i := range t.Splits {
		if !t.Splits[i].Valid() {
			return false
		}
	}
	return true
}
