package handlers_test

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
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

func findSecurity(client *http.Client, symbol string, tipe handlers.SecurityType) (*handlers.Security, error) {
	securities, err := getSecurities(client)
	if err != nil {
		return nil, err
	}
	for _, security := range *securities.Securities {
		if security.Symbol == symbol && security.Type == tipe {
			return security, nil
		}
	}
	return nil, fmt.Errorf("Unable to find security: \"%s\"", symbol)
}

func findAccount(client *http.Client, name string, tipe handlers.AccountType, securityid int64) (*handlers.Account, error) {
	accounts, err := getAccounts(client)
	if err != nil {
		return nil, err
	}
	for _, account := range *accounts.Accounts {
		if account.Name == name && account.Type == tipe && account.SecurityId == securityid {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("Unable to find account: \"%s\"", name)
}

func TestImportOFX401kMutualFunds(t *testing.T) {
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

		account := &handlers.Account{
			SecurityId:      d.securities[0].SecurityId,
			UserId:          d.users[0].UserId,
			ParentAccountId: -1,
			Type:            handlers.Investment,
			Name:            "401k",
		}

		account, err = createAccount(d.clients[0], account)
		if err != nil {
			t.Fatalf("Error creating 401k account: %s\n", err)
		}

		// Import and ensure it didn't return a nasty error code
		if err = importOFX(d.clients[0], account.AccountId, "handlers_testdata/401k_mutualfunds.ofx"); err != nil {
			t.Fatalf("Error importing OFX: %s\n", err)
		}
		accountBalanceHelper(t, d.clients[0], account, "-192.10")

		// Make sure the security was created and that the trading account has
		// the right value
		security, err := findSecurity(d.clients[0], "VANGUARD TARGET 2045", handlers.Stock)
		if err != nil {
			t.Fatalf("Error finding VANGUARD TARGET 2045 security: %s\n", err)
		}
		tradingaccount, err := findAccount(d.clients[0], "VANGUARD TARGET 2045", handlers.Trading, security.SecurityId)
		if err != nil {
			t.Fatalf("Error finding VANGUARD TARGET 2045 trading account: %s\n", err)
		}
		accountBalanceHelper(t, d.clients[0], tradingaccount, "-3.35400")

		// Ensure actual holding account was created and in the correct place
		investmentaccount, err := findAccount(d.clients[0], "VANGUARD TARGET 2045", handlers.Investment, security.SecurityId)
		if err != nil {
			t.Fatalf("Error finding VANGUARD TARGET 2045 investment account: %s\n", err)
		}
		accountBalanceHelper(t, d.clients[0], investmentaccount, "3.35400")
		if investmentaccount.ParentAccountId != account.AccountId {
			t.Errorf("Expected imported security account to be child of investment account it's imported into\n")
		}
	})
}
