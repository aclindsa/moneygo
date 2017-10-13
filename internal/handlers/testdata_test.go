package handlers_test

import (
	"encoding/json"
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strings"
	"testing"
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

	for i, account := range t.accounts {
		account.SecurityId = t2.securities[t.accounts[i].SecurityId].SecurityId
		if account.ParentAccountId != -1 {
			account.ParentAccountId = t2.accounts[t.accounts[i].ParentAccountId].AccountId
		}
		a2, err := createAccount(t2.clients[account.UserId], &account)
		if err != nil {
			return nil, err
		}
		t2.accounts = append(t2.accounts, *a2)
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
		},
	},
}
