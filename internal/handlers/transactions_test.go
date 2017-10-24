package handlers_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func createTransaction(client *http.Client, transaction *handlers.Transaction) (*handlers.Transaction, error) {
	var s handlers.Transaction
	err := create(client, transaction, &s, "/transaction/", "transaction")
	return &s, err
}

func getTransaction(client *http.Client, transactionid int64) (*handlers.Transaction, error) {
	var s handlers.Transaction
	err := read(client, &s, "/transaction/"+strconv.FormatInt(transactionid, 10), "transaction")
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func getTransactions(client *http.Client) (*handlers.TransactionList, error) {
	var tl handlers.TransactionList
	err := read(client, &tl, "/transaction/", "transactions")
	if err != nil {
		return nil, err
	}
	return &tl, nil
}

func updateTransaction(client *http.Client, transaction *handlers.Transaction) (*handlers.Transaction, error) {
	var s handlers.Transaction
	err := update(client, transaction, &s, "/transaction/"+strconv.FormatInt(transaction.TransactionId, 10), "transaction")
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func deleteTransaction(client *http.Client, s *handlers.Transaction) error {
	err := remove(client, "/transaction/"+strconv.FormatInt(s.TransactionId, 10), "transaction")
	if err != nil {
		return err
	}
	return nil
}

func ensureTransactionsMatch(t *testing.T, expected, tran *handlers.Transaction, accounts *[]handlers.Account, matchtransactionids, matchsplitids bool) {
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
	if tran.Date != expected.Date {
		t.Errorf("Date doesn't match")
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
				origsplit.Amount == s.Amount &&
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

func TestCreateTransaction(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i, orig := range data[0].transactions {
			transaction := d.transactions[i]

			ensureTransactionsMatch(t, &orig, &transaction, &d.accounts, false, false)
		}

		// Don't allow imbalanced transactions
		tran := handlers.Transaction{
			UserId:      d.users[0].UserId,
			Description: "Imbalanced",
			Date:        time.Date(2017, time.September, 1, 0, 00, 00, 0, time.UTC),
			Splits: []*handlers.Split{
				&handlers.Split{
					Status:     handlers.Reconciled,
					AccountId:  d.accounts[1].AccountId,
					SecurityId: -1,
					Amount:     "-39.98",
				},
				&handlers.Split{
					Status:     handlers.Entered,
					AccountId:  d.accounts[4].AccountId,
					SecurityId: -1,
					Amount:     "39.99",
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
		tran.Splits = []*handlers.Split{}
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

			tran, err := updateTransaction(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error updating transaction: %s\n", err)
			}

			ensureTransactionsMatch(t, &curr, tran, nil, true, true)

			tran.Splits = []*handlers.Split{}
			for _, s := range curr.Splits {
				var split handlers.Split
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
			tran.Splits[len(tran.Splits)-1].Amount = "42"
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
			tran.Splits = []*handlers.Split{}
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
		for i := 1; i < len(data[0].transactions); i++ {
			orig := data[0].transactions[i]
			curr := d.transactions[i]

			err := deleteTransaction(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error deleting transaction: %s\n", err)
			}

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
