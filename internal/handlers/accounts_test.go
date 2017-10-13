package handlers_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
)

func createAccount(client *http.Client, account *handlers.Account) (*handlers.Account, error) {
	var a handlers.Account
	err := create(client, account, &a, "/account/", "account")
	return &a, err
}

func getAccount(client *http.Client, accountid int64) (*handlers.Account, error) {
	var a handlers.Account
	err := read(client, &a, "/account/"+strconv.FormatInt(accountid, 10), "account")
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func updateAccount(client *http.Client, account *handlers.Account) (*handlers.Account, error) {
	var a handlers.Account
	err := update(client, account, &a, "/account/"+strconv.FormatInt(account.AccountId, 10), "account")
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func deleteAccount(client *http.Client, a *handlers.Account) error {
	err := remove(client, "/account/"+strconv.FormatInt(a.AccountId, 10), "account")
	if err != nil {
		return err
	}
	return nil
}

func TestCreateAccount(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			a := d.accounts[i]

			if a.AccountId == 0 {
				t.Errorf("Unable to create security: %+v", a)
			}
			if a.Type != orig.Type {
				t.Errorf("Type doesn't match")
			}
			if a.Name != orig.Name {
				t.Errorf("Name doesn't match")
			}
		}
	})
}

func TestGetAccount(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			curr := d.accounts[i]

			a, err := getAccount(d.clients[orig.UserId], curr.AccountId)
			if err != nil {
				t.Fatalf("Error fetching accounts: %s\n", err)
			}
			if a.SecurityId != curr.SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if a.Type != orig.Type {
				t.Errorf("Type doesn't match")
			}
			if a.Name != orig.Name {
				t.Errorf("Name doesn't match")
			}
		}
	})
}

func TestUpdateAccount(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			curr := d.accounts[i]

			curr.Name = "blah"
			curr.Type = handlers.Payable
			curr.SecurityId = d.securities[1].SecurityId

			a, err := updateAccount(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error updating account: %s\n", err)
			}

			if a.AccountId != curr.AccountId {
				t.Errorf("AccountId doesn't match")
			}
			if a.Type != curr.Type {
				t.Errorf("Type doesn't match")
			}
			if a.Name != curr.Name {
				t.Errorf("Name doesn't match")
			}
			if a.SecurityId != curr.SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
		}

		orig := data[0].accounts[0]
		curr := d.accounts[0]
		curr.ParentAccountId = curr.AccountId

		a, err := updateAccount(d.clients[orig.UserId], &curr)
		if err == nil {
			t.Fatalf("Expected error updating account to be own parent: %+v\n", a)
		}

		orig = data[0].accounts[0]
		curr = d.accounts[0]
		curr.ParentAccountId = 999999

		a, err = updateAccount(d.clients[orig.UserId], &curr)
		if err == nil {
			t.Fatalf("Expected error updating account with invalid parent: %+v\n", a)
		}

		orig = data[0].accounts[0]
		curr = d.accounts[0]
		child := d.accounts[1]
		curr.ParentAccountId = child.AccountId

		a, err = updateAccount(d.clients[orig.UserId], &curr)
		if err == nil {
			t.Fatalf("Expected error updating account with circular parent relationship: %+v\n", a)
		}
	})
}

func TestDeleteAccount(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			curr := d.accounts[i]

			err := deleteAccount(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error deleting account: %s\n", err)
			}

			_, err = getAccount(d.clients[orig.UserId], curr.AccountId)
			if err == nil {
				t.Fatalf("Expected error fetching deleted account")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error fetching deleted account: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error fetching deleted account")
			}
		}
	})
}
