package handlers_test

import (
	"net/http"
	"strconv"
	"testing"
)

func importOFX(client *http.Client, accountid int64, filename string) error {
	return uploadFile(client, filename, "/v1/accounts/"+strconv.FormatInt(accountid, 10)+"/imports/ofxfile")
}

func TestImportOFXChecking(t *testing.T) {
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
		if err = importOFX(d.clients[0], d.accounts[1].AccountId, "handlers_testdata/checking_20171126.ofx"); err != nil {
			t.Fatalf("Error importing OFX: %s\n", err)
		}
		accountBalanceHelper(t, d.clients[0], &d.accounts[1], "2493.19")

		if err = importOFX(d.clients[0], d.accounts[1].AccountId, "handlers_testdata/checking_20171129.ofx"); err != nil {
			t.Fatalf("Error importing OFX: %s\n", err)
		}
		accountBalanceHelper(t, d.clients[0], &d.accounts[1], "5336.27")
	})
}

func TestImportOFXCreditCard(t *testing.T) {
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
		if err = importOFX(d.clients[0], d.accounts[7].AccountId, "handlers_testdata/creditcard.ofx"); err != nil {
			t.Fatalf("Error importing OFX: %s\n", err)
		}
		accountBalanceHelper(t, d.clients[0], &d.accounts[7], "-4.49")
	})
}
