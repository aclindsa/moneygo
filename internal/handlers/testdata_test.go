package handlers_test

import (
	"encoding/json"
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Needed because handlers.User doesn't allow Password to be written to JSON
type User struct {
	UserId          int64
	DefaultCurrency int64 // SecurityId of default currency, or ISO4217 code for it if creating new user
	Name            string
	Username        string
	Password        string
	PasswordHash    string
	Email           string
}

func (u *User) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(u)
}

func (u *User) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(u)
}

// TestData
type TestData struct {
	initialized  bool
	users        []User
	clients      []*http.Client
	securities   []handlers.Security
	accounts     []handlers.Account // accounts must appear after their parents in this slice
	transactions []handlers.Transaction
	prices       []handlers.Price
	reports      []handlers.Report
}

type TestDataFunc func(*testing.T, *TestData)

func (t *TestData) initUser(user *User, userid int) error {
	newuser, err := createUser(user)
	if err != nil {
		return err
	}
	t.users = append(t.users, *newuser)

	// make a copy of the user so we can set the password for creating the
	// session without disturbing the original
	userWithPassword := *newuser
	userWithPassword.Password = user.Password

	client, err := newSession(&userWithPassword)
	if err != nil {
		return err
	}

	t.clients = append(t.clients, client)

	// TODO initialize everything else owned by this user in the TestData struct

	return nil
}

// Initialize makes requests to the server to create all of the objects
// represented in it before returning a copy of the data, with all of the *Id
// fields updated to their actual values
func (t *TestData) Initialize() (*TestData, error) {
	var t2 TestData
	for userid, user := range t.users {
		err := t2.initUser(&user, userid)
		if err != nil {
			return nil, err
		}
	}

	for _, security := range t.securities {
		s2, err := createSecurity(t2.clients[security.UserId], &security)
		if err != nil {
			return nil, err
		}
		t2.securities = append(t2.securities, *s2)
	}

	for _, account := range t.accounts {
		account.SecurityId = t2.securities[account.SecurityId].SecurityId
		if account.ParentAccountId != -1 {
			account.ParentAccountId = t2.accounts[account.ParentAccountId].AccountId
		}
		a2, err := createAccount(t2.clients[account.UserId], &account)
		if err != nil {
			return nil, err
		}
		t2.accounts = append(t2.accounts, *a2)
	}

	for i, transaction := range t.transactions {
		transaction.Splits = []*handlers.Split{}
		for _, s := range t.transactions[i].Splits {
			// Make a copy of the split since Splits is a slice of pointers so
			// copying the transaction doesn't
			split := *s
			split.AccountId = t2.accounts[split.AccountId].AccountId
			transaction.Splits = append(transaction.Splits, &split)
		}
		tt2, err := createTransaction(t2.clients[transaction.UserId], &transaction)
		if err != nil {
			return nil, err
		}
		t2.transactions = append(t2.transactions, *tt2)
	}

	t2.initialized = true
	return &t2, nil
}

func (t *TestData) Teardown() error {
	if !t.initialized {
		return fmt.Errorf("Cannot teardown uninitialized TestData")
	}
	for userid, user := range t.users {
		err := deleteUser(t.clients[userid], &user)
		if err != nil {
			return err
		}
	}
	return nil
}

var data = []TestData{
	{
		users: []User{
			User{
				DefaultCurrency: 840, // USD
				Name:            "John Smith",
				Username:        "jsmith",
				Password:        "hunter2",
				Email:           "jsmith@example.com",
			},
			User{
				DefaultCurrency: 978, // Euro
				Name:            "Billy Bob",
				Username:        "bbob6",
				Password:        "#)$&!KF(*ADAHK#@*(FAJSDkalsdf98af32klhf98sd8a'2938LKJD",
				Email:           "bbob+moneygo@my-domain.com",
			},
		},
		securities: []handlers.Security{
			handlers.Security{
				UserId:      0,
				Name:        "USD",
				Description: "US Dollar",
				Symbol:      "$",
				Precision:   2,
				Type:        handlers.Currency,
				AlternateId: "840",
			},
			handlers.Security{
				UserId:      0,
				Name:        "SPY",
				Description: "SPDR S&P 500 ETF Trust",
				Symbol:      "SPY",
				Precision:   5,
				Type:        handlers.Stock,
				AlternateId: "78462F103",
			},
			handlers.Security{
				UserId:      1,
				Name:        "EUR",
				Description: "Euro",
				Symbol:      "€",
				Precision:   2,
				Type:        handlers.Currency,
				AlternateId: "978",
			},
		},
		accounts: []handlers.Account{
			handlers.Account{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: -1,
				Type:            handlers.Asset,
				Name:            "Assets",
			},
			handlers.Account{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: 0,
				Type:            handlers.Asset,
				Name:            "Credit Union Checking",
			},
			handlers.Account{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: -1,
				Type:            handlers.Expense,
				Name:            "Expenses",
			},
			handlers.Account{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: 2,
				Type:            handlers.Expense,
				Name:            "Groceries",
			},
			handlers.Account{
				UserId:          0,
				SecurityId:      0,
				ParentAccountId: 2,
				Type:            handlers.Expense,
				Name:            "Cable",
			},
			handlers.Account{
				UserId:          1,
				SecurityId:      2,
				ParentAccountId: -1,
				Type:            handlers.Asset,
				Name:            "Assets",
			},
			handlers.Account{
				UserId:          1,
				SecurityId:      2,
				ParentAccountId: -1,
				Type:            handlers.Expense,
				Name:            "Expenses",
			},
		},
		transactions: []handlers.Transaction{
			handlers.Transaction{
				UserId:      0,
				Description: "weekly groceries",
				Date:        time.Date(2017, time.October, 15, 1, 16, 59, 0, time.UTC),
				Splits: []*handlers.Split{
					&handlers.Split{
						Status:     handlers.Reconciled,
						AccountId:  1,
						SecurityId: -1,
						Amount:     "-5.6",
					},
					&handlers.Split{
						Status:     handlers.Reconciled,
						AccountId:  3,
						SecurityId: -1,
						Amount:     "5.6",
					},
				},
			},
			handlers.Transaction{
				UserId:      0,
				Description: "Cable",
				Date:        time.Date(2017, time.September, 1, 0, 00, 00, 0, time.UTC),
				Splits: []*handlers.Split{
					&handlers.Split{
						Status:     handlers.Reconciled,
						AccountId:  1,
						SecurityId: -1,
						Amount:     "-39.99",
					},
					&handlers.Split{
						Status:     handlers.Entered,
						AccountId:  4,
						SecurityId: -1,
						Amount:     "39.99",
					},
				},
			},
			handlers.Transaction{
				UserId:      1,
				Description: "Gas",
				Date:        time.Date(2017, time.November, 1, 13, 19, 50, 0, time.UTC),
				Splits: []*handlers.Split{
					&handlers.Split{
						Status:     handlers.Reconciled,
						AccountId:  5,
						SecurityId: -1,
						Amount:     "-24.56",
					},
					&handlers.Split{
						Status:     handlers.Entered,
						AccountId:  6,
						SecurityId: -1,
						Amount:     "24.56",
					},
				},
			},
		},
	},
}
