package integration_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"strconv"
	"testing"
)

func createAccount(client *http.Client, account *models.Account) (*models.Account, error) {
	var a models.Account
	err := create(client, account, &a, "/v1/accounts/")
	return &a, err
}

func getAccount(client *http.Client, accountid int64) (*models.Account, error) {
	var a models.Account
	err := read(client, &a, "/v1/accounts/"+strconv.FormatInt(accountid, 10))
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func getAccounts(client *http.Client) (*models.AccountList, error) {
	var al models.AccountList
	err := read(client, &al, "/v1/accounts/")
	if err != nil {
		return nil, err
	}
	return &al, nil
}

func updateAccount(client *http.Client, account *models.Account) (*models.Account, error) {
	var a models.Account
	err := update(client, account, &a, "/v1/accounts/"+strconv.FormatInt(account.AccountId, 10))
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func deleteAccount(client *http.Client, a *models.Account) error {
	err := remove(client, "/v1/accounts/"+strconv.FormatInt(a.AccountId, 10))
	if err != nil {
		return err
	}
	return nil
}

func TestCreateAccount(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			a := d.accounts[i]

			if a.AccountId == 0 {
				t.Errorf("Unable to create account: %+v", a)
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
		for i := 0; i < len(data[0].accounts); i++ {
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

func TestGetAccounts(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		al, err := getAccounts(d.clients[0])
		if err != nil {
			t.Fatalf("Error fetching accounts: %s\n", err)
		}

		numaccounts := 0
		foundIds := make(map[int64]bool)
		for i := 0; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			curr := d.accounts[i]

			if curr.UserId != d.users[0].UserId {
				continue
			}
			numaccounts += 1

			found := false
			for _, a := range *al.Accounts {
				if orig.Name == a.Name && orig.Type == a.Type && a.ExternalAccountId == orig.ExternalAccountId && d.securities[orig.SecurityId].SecurityId == a.SecurityId && ((orig.ParentAccountId == -1 && a.ParentAccountId == -1) || d.accounts[orig.ParentAccountId].AccountId == a.ParentAccountId) {
					if _, ok := foundIds[a.AccountId]; ok {
						continue
					}
					foundIds[a.AccountId] = true
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unable to find matching account: %+v", orig)
			}
		}

		if numaccounts != len(*al.Accounts) {
			t.Fatalf("Expected %d accounts, received %d", numaccounts, len(*al.Accounts))
		}
	})
}

func TestUpdateAccount(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].accounts); i++ {
			orig := data[0].accounts[i]
			curr := d.accounts[i]

			curr.Name = "blah"
			curr.Type = models.Payable
			for _, s := range d.securities {
				if s.UserId == curr.UserId {
					curr.SecurityId = s.SecurityId
					break
				}
			}

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
		for i := 0; i < len(data[0].accounts); i++ {
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
