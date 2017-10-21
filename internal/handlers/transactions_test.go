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

			// Make sure we can't create an unbalanced transaction
			tran.Splits = []*handlers.Split{}
			for _, s := range curr.Splits {
				var split handlers.Split
				split = *s
				tran.Splits = append(tran.Splits, &split)
			}
			tran.Splits[len(tran.Splits)-1].Amount = "42"
			_, err = updateTransaction(d.clients[orig.UserId], tran)
			if err == nil {
				t.Fatalf("Expected error updating imbalanced splits")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error updating imbalanced splits: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error updating imbalanced splits")
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
