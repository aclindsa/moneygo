package main

type AccountType int64

const (
	Bank       AccountType = 1
	Cash                   = 2
	Asset                  = 3
	Liability              = 4
	Investment             = 5
	Income                 = 6
	Expense                = 7
)

type Account struct {
	AccountId  int64
	UserId     int64
	SecurityId int64
	Type       AccountType
	Name       string
}
