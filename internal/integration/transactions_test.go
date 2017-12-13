package integration_test

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func createTransaction(client *http.Client, transaction *models.Transaction) (*models.Transaction, error) {
	var s models.Transaction
	err := create(client, transaction, &s, "/v1/transactions/")
	return &s, err
}

func getTransaction(client *http.Client, transactionid int64) (*models.Transaction, error) {
	var s models.Transaction
	err := read(client, &s, "/v1/transactions/"+strconv.FormatInt(transactionid, 10))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func getTransactions(client *http.Client) (*models.TransactionList, error) {
	var tl models.TransactionList
	err := read(client, &tl, "/v1/transactions/")
	if err != nil {
		return nil, err
	}
	return &tl, nil
}

func getAccountTransactions(client *http.Client, accountid, page, limit int64, sort string) (*models.AccountTransactionsList, error) {
	var atl models.AccountTransactionsList
	params := url.Values{}

	query := fmt.Sprintf("/v1/accounts/%d/transactions/", accountid)
	if page != 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if limit != 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if len(sort) != 0 {
		params.Set("sort", sort)
		query += "?" + params.Encode()
	}

	err := read(client, &atl, query)
	if err != nil {
		return nil, err
	}
	return &atl, nil
}

func updateTransaction(client *http.Client, transaction *models.Transaction) (*models.Transaction, error) {
	var s models.Transaction
	err := update(client, transaction, &s, "/v1/transactions/"+strconv.FormatInt(transaction.TransactionId, 10))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func deleteTransaction(client *http.Client, s *models.Transaction) error {
	err := remove(client, "/v1/transactions/"+strconv.FormatInt(s.TransactionId, 10))
	if err != nil {
		return err
	}
	return nil
}

func ensureTransactionsMatch(t *testing.T, expected, tran *models.Transaction, accounts *[]models.Account, matchtransactionids, matchsplitids bool) {
	t.Helper()

	if tran.TransactionId == 0 {
		t.Errorf("TransactionId is 0")
	}

	if matchtransactionids && tran.TransactionId != expected.TransactionId {
		t.Errorf("TransactionId (%d) doesn't match what's expected (%d)", tran.TransactionId, expected.TransactionId)
	}
	if tran.Description != expected.Description {
		t.Errorf("Description doesn't match")
	}
	if !tran.Date.Equal(expected.Date) {
		t.Errorf("Date (%+v) differs from expected (%+v)", tran.Date, expected.Date)
	}

	if len(tran.Splits) != len(expected.Splits) {
		t.Fatalf("Expected %d splits, received %d", len(expected.Splits), len(tran.Splits))
	}

	foundIds := make(map[int64]bool)
	for j := 0; j < len(expected.Splits); j++ {
		origsplit := expected.Splits[j]

		if tran.Splits[j].TransactionId != tran.TransactionId {
			t.Fatalf("Split TransactionId doesn't match transaction's")
		}

		found := false
		for _, s := range tran.Splits {
			if s.SplitId == 0 {
				t.Errorf("Found SplitId that's 0")
			}
			accountid := origsplit.AccountId
			if accounts != nil {
				accountid = (*accounts)[accountid].AccountId
			}
			if origsplit.Status == s.Status &&
				origsplit.ImportSplitType == s.ImportSplitType &&
				s.AccountId == accountid &&
				s.SecurityId == -1 &&
				origsplit.RemoteId == origsplit.RemoteId &&
				origsplit.Number == s.Number &&
				origsplit.Memo == s.Memo &&
				origsplit.Amount.Cmp(&s.Amount.Rat) == 0 &&
				(!matchsplitids || origsplit.SplitId == s.SplitId) {

				if _, ok := foundIds[s.SplitId]; ok {
					continue
				}
				foundIds[s.SplitId] = true
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unable to find matching split: %+v", origsplit)
		}
	}
}

func getAccountVersionMap(t *testing.T, client *http.Client, tran *models.Transaction) map[int64]*models.Account {
	t.Helper()
	accountMap := make(map[int64]*models.Account)
	for _, split := range tran.Splits {
		account, err := getAccount(client, split.AccountId)
		if err != nil {
			t.Fatalf("Error fetching split's account while updating transaction: %s\n", err)
		}
		accountMap[account.AccountId] = account
	}
	return accountMap
}

func checkAccountVersionsUpdated(t *testing.T, client *http.Client, accountMap map[int64]*models.Account, tran *models.Transaction) {
	for _, split := range tran.Splits {
		account, err := getAccount(client, split.AccountId)
		if err != nil {
			t.Fatalf("Error fetching split's account after updating transaction: %s\n", err)
		}
		if account.AccountVersion <= accountMap[split.AccountId].AccountVersion {
			t.Errorf("Failed to update account version when updating transaction split\n")
		}
	}
}

func TestCreateTransaction(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i, orig := range data[0].transactions {
			transaction := d.transactions[i]

			ensureTransactionsMatch(t, &orig, &transaction, &d.accounts, false, false)

			accountMap := getAccountVersionMap(t, d.clients[orig.UserId], &transaction)
			_, err := createTransaction(d.clients[orig.UserId], &transaction)
			if err != nil {
				t.Fatalf("Unxpected error creating transaction")
			}
			checkAccountVersionsUpdated(t, d.clients[orig.UserId], accountMap, &transaction)
		}

		// Don't allow imbalanced transactions
		tran := models.Transaction{
			UserId:      d.users[0].UserId,
			Description: "Imbalanced",
			Date:        time.Date(2017, time.September, 1, 0, 00, 00, 0, time.UTC),
			Splits: []*models.Split{
				{
					Status:     models.Reconciled,
					AccountId:  d.accounts[1].AccountId,
					SecurityId: -1,
					Amount:     NewAmount("-39.98"),
				},
				{
					Status:     models.Entered,
					AccountId:  d.accounts[4].AccountId,
					SecurityId: -1,
					Amount:     NewAmount("39.99"),
				},
			},
		}
		_, err := createTransaction(d.clients[0], &tran)
		if err == nil {
			t.Fatalf("Expected error creating imbalanced transaction")
		}
		if herr, ok := err.(*handlers.Error); ok {
			if herr.ErrorId != 3 { // Invalid requeset
				t.Fatalf("Unexpected API error creating imbalanced transaction: %s", herr)
			}
		} else {
			t.Fatalf("Unexpected error creating imbalanced transaction")
		}

		// Don't allow transactions with 0 splits
		tran.Splits = []*models.Split{}
		_, err = createTransaction(d.clients[0], &tran)
		if err == nil {
			t.Fatalf("Expected error creating with zero splits")
		}
		if herr, ok := err.(*handlers.Error); ok {
			if herr.ErrorId != 3 { // Invalid requeset
				t.Fatalf("Unexpected API error creating with zero splits: %s", herr)
			}
		} else {
			t.Fatalf("Unexpected error creating zero splits")
		}

		// Don't allow creating a transaction for another user
		tran.UserId = d.users[1].UserId
		_, err = createTransaction(d.clients[0], &tran)
		if err == nil {
			t.Fatalf("Expected error creating transaction for another user")
		}
		if herr, ok := err.(*handlers.Error); ok {
			if herr.ErrorId != 3 { // Invalid request
				t.Fatalf("Unexpected API error creating transction for another user: %s", herr)
			}
		} else {
			t.Fatalf("Unexpected error creating transaction for another user")
		}
	})
}

func TestGetTransaction(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].transactions); i++ {
			orig := data[0].transactions[i]
			curr := d.transactions[i]

			tran, err := getTransaction(d.clients[orig.UserId], curr.TransactionId)
			if err != nil {
				t.Fatalf("Error fetching transaction: %s\n", err)
			}

			ensureTransactionsMatch(t, &curr, tran, nil, true, true)
		}
	})
}

func TestGetTransactions(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		tl, err := getTransactions(d.clients[0])
		if err != nil {
			t.Fatalf("Error fetching transactions: %s\n", err)
		}

		numtransactions := 0
		foundIds := make(map[int64]bool)
		for i := 0; i < len(data[0].transactions); i++ {
			orig := data[0].transactions[i]
			curr := d.transactions[i]

			if curr.UserId != d.users[0].UserId {
				continue
			}
			numtransactions += 1

			found := false
			for _, tran := range *tl.Transactions {
				if tran.TransactionId == curr.TransactionId {
					ensureTransactionsMatch(t, &curr, tran, nil, true, true)
					if _, ok := foundIds[tran.TransactionId]; ok {
						continue
					}
					foundIds[tran.TransactionId] = true
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unable to find matching transaction: %+v", orig)
			}
		}

		if numtransactions != len(*tl.Transactions) {
			t.Fatalf("Expected %d transactions, received %d", numtransactions, len(*tl.Transactions))
		}
	})
}

func TestUpdateTransaction(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].transactions); i++ {
			orig := data[0].transactions[i]
			curr := d.transactions[i]

			curr.Description = "more money"
			curr.Date = time.Date(2017, time.October, 18, 10, 41, 40, 0, time.UTC)

			accountMap := getAccountVersionMap(t, d.clients[orig.UserId], &curr)

			tran, err := updateTransaction(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error updating transaction: %s\n", err)
			}

			checkAccountVersionsUpdated(t, d.clients[orig.UserId], accountMap, tran)
			checkAccountVersionsUpdated(t, d.clients[orig.UserId], accountMap, &curr)

			ensureTransactionsMatch(t, &curr, tran, nil, true, true)

			tran.Splits = []*models.Split{}
			for _, s := range curr.Splits {
				var split models.Split
				split = *s
				tran.Splits = append(tran.Splits, &split)
			}

			// Don't allow updating transactions for other/invalid users
			tran.UserId = tran.UserId + 1
			tran2, err := updateTransaction(d.clients[orig.UserId], tran)
			if tran2.UserId != curr.UserId {
				t.Fatalf("Allowed updating transaction to have wrong UserId\n")
			}
			tran.UserId = curr.UserId

			// Make sure we can't create an unbalanced transaction
			tran.Splits[len(tran.Splits)-1].Amount = NewAmount("42")
			_, err = updateTransaction(d.clients[orig.UserId], tran)
			if err == nil {
				t.Fatalf("Expected error updating imbalanced transaction")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error updating imbalanced transaction: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error updating imbalanced transaction")
			}

			// Don't allow transactions with 0 splits
			tran.Splits = []*models.Split{}
			_, err = updateTransaction(d.clients[orig.UserId], tran)
			if err == nil {
				t.Fatalf("Expected error updating with zero splits")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error updating with zero splits: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error updating zero splits")
			}
		}
	})
}

func TestDeleteTransaction(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].transactions); i++ {
			orig := data[0].transactions[i]
			curr := d.transactions[i]

			accountMap := getAccountVersionMap(t, d.clients[orig.UserId], &curr)

			err := deleteTransaction(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error deleting transaction: %s\n", err)
			}
			checkAccountVersionsUpdated(t, d.clients[orig.UserId], accountMap, &curr)

			_, err = getTransaction(d.clients[orig.UserId], curr.TransactionId)
			if err == nil {
				t.Fatalf("Expected error fetching deleted transaction")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error fetching deleted transaction: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error fetching deleted transaction")
			}
		}
	})
}

func helperTestAccountTransactions(t *testing.T, d *TestData, account *models.Account, limit int64, sort string) {
	if account.UserId != d.users[0].UserId {
		return
	}

	var transactions []models.Transaction
	var lastFetchCount int64

	for page := int64(0); page == 0 || lastFetchCount > 0; page++ {
		atl, err := getAccountTransactions(d.clients[0], account.AccountId, page, limit, sort)
		if err != nil {
			t.Fatalf("Error fetching account transactions: %s\n", err)
		}
		if limit != 0 && atl.Transactions != nil && int64(len(*atl.Transactions)) > limit {
			t.Errorf("Exceeded limit of %d transactions (returned %d)\n", limit, len(*atl.Transactions))
		}
		if atl.Transactions != nil {
			for _, tran := range *atl.Transactions {
				transactions = append(transactions, *tran)
			}
			lastFetchCount = int64(len(*atl.Transactions))
		} else {
			lastFetchCount = -1
		}
	}

	var lastDate time.Time
	for _, tran := range transactions {
		if lastDate.IsZero() {
			lastDate = tran.Date
			continue
		} else if sort == "date-desc" && lastDate.Before(tran.Date) {
			t.Errorf("Sorted by date-desc, but later transaction has later date")
		} else if sort == "date-asc" && lastDate.After(tran.Date) {
			t.Errorf("Sorted by date-asc, but later transaction has earlier date")
		}
		lastDate = tran.Date
	}

	numtransactions := 0
	foundIds := make(map[int64]bool)
	for i := 0; i < len(d.transactions); i++ {
		curr := d.transactions[i]

		if curr.UserId != d.users[0].UserId {
			continue
		}

		// Don't consider this transaction if we didn't find a split
		// for the account we're considering
		account_found := false
		for _, s := range curr.Splits {
			if s.AccountId == account.AccountId {
				account_found = true
				break
			}
		}
		if !account_found {
			continue
		}

		numtransactions += 1

		found := false
		for _, tran := range transactions {
			if tran.TransactionId == curr.TransactionId {
				ensureTransactionsMatch(t, &curr, &tran, nil, true, true)
				if _, ok := foundIds[tran.TransactionId]; ok {
					continue
				}
				foundIds[tran.TransactionId] = true
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unable to find matching transaction: %+v", curr)
			t.Errorf("Transactions: %+v\n", transactions)
		}
	}

	if numtransactions != len(transactions) {
		t.Fatalf("Expected %d transactions, received %d", numtransactions, len(transactions))
	}
}

func TestAccountTransactions(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for _, account := range d.accounts {
			helperTestAccountTransactions(t, d, &account, 0, "date-desc")
			helperTestAccountTransactions(t, d, &account, 0, "date-asc")
			helperTestAccountTransactions(t, d, &account, 1, "date-desc")
			helperTestAccountTransactions(t, d, &account, 1, "date-asc")
			helperTestAccountTransactions(t, d, &account, 2, "date-desc")
			helperTestAccountTransactions(t, d, &account, 2, "date-asc")
		}
	})
}
