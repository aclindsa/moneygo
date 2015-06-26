package main

import (
	"math/big"
	"time"
)

type Split struct {
	SplitId       int64
	TransactionId int64
	AccountId     int64
	Number        int64 // Check or reference number
	Memo          string
	Amount        big.Rat
	Debit         bool
}

type TransactionStatus int64

const (
	Entered    TransactionStatus = 1
	Cleared                      = 2
	Reconciled                   = 3
	Voided                       = 4
)

type Transaction struct {
	TransactionId int64
	UserId        int64
	Description   string
	Status        TransactionStatus
	Date          time.Time
}
