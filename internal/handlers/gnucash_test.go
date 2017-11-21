package handlers_test

import (
	"bytes"
	"github.com/aclindsa/moneygo/internal/handlers"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
)

func importGnucash(client *http.Client, filename string) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	filewriter, err := mw.CreateFormFile("gnucash", filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(filewriter, file); err != nil {
		return err
	}

	mw.Close()

	response, err := client.Post(server.URL+"/v1/imports/gnucash", mw.FormDataContentType(), &buf)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return err
	}

	var e handlers.Error
	err = (&e).Read(string(body))
	if err != nil {
		return err
	}
	if e.ErrorId != 0 || len(e.ErrorString) != 0 {
		return &e
	}

	return nil
}

func TestImportGnucash(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		// Ensure there's only one USD currency
		oldDefault, err := getSecurity(d.clients[0], d.users[0].DefaultCurrency)
		if err != nil {
			t.Fatalf("Error fetching default security: %s\n", err)
		}
		d.users[0].DefaultCurrency = d.securities[0].SecurityId
		if _, err := updateUser(d.clients[0], &d.users[0]); err != nil {
			t.Fatalf("Error updating user: %s\n", err)
		}
		if err := deleteSecurity(d.clients[0], oldDefault); err != nil {
			t.Fatalf("Error removing default security: %s\n", err)
		}

		// Import and ensure it didn't return a nasty error code
		if err = importGnucash(d.clients[0], "handlers_testdata/example.gnucash"); err != nil {
			t.Fatalf("Error importing from Gnucash: %s\n", err)
		}

		// Next, find the Expenses/Groceries account
		var groceries *handlers.Account
		accounts, err := getAccounts(d.clients[0])
		if err != nil {
			t.Fatalf("Error fetching accounts: %s\n", err)
		}
		for _, account := range *accounts.Accounts {
			if account.Name == "Groceries" {
				groceries = &account
				break
			}
		}
		if groceries == nil {
			t.Fatalf("Couldn't find 'Expenses/Groceries' account")
		}

		grocerytransactions, err := getAccountTransactions(d.clients[0], groceries.AccountId, 0, 0, "")
		if err != nil {
			t.Fatalf("Couldn't fetch account transactions for 'Expenses/Groceries': %s\n", err)
		}

		// 87.19 from preexisting transactions and 200.37 from Gnucash
		if grocerytransactions.EndingBalance != "287.56" {
			t.Errorf("Expected ending balance for 'Expenses/Groceries' to be '287.56', but found %s\n", grocerytransactions.EndingBalance)
		}
	})
}
