package handlers_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
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

func TestCreateTransaction(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i, orig := range data[0].transactions {
			transaction := d.transactions[i]

			if transaction.TransactionId == 0 {
				t.Errorf("Unable to create transaction: %+v", transaction)
			}
			if transaction.Description != orig.Description {
				t.Errorf("Description doesn't match")
			}
			if transaction.Date != orig.Date {
				t.Errorf("Date doesn't match")
			}

			if len(transaction.Splits) != len(orig.Splits) {
				t.Fatalf("Expected %d splits, received %d", len(orig.Splits), len(transaction.Splits))
			}

			foundIds := make(map[int64]bool)
			for j := 0; j < len(orig.Splits); j++ {
				origsplit := orig.Splits[j]

				if transaction.Splits[j].TransactionId != transaction.TransactionId {
					t.Fatalf("Split TransactionId doesn't match transaction's")
				}

				found := false
				for _, s := range transaction.Splits {
					if origsplit.Status == s.Status && origsplit.ImportSplitType == s.ImportSplitType && s.AccountId == d.accounts[origsplit.AccountId].AccountId && s.SecurityId == -1 && origsplit.RemoteId == origsplit.RemoteId && origsplit.Number == s.Number && origsplit.Memo == s.Memo && origsplit.Amount == s.Amount {
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
			if tran.TransactionId != curr.TransactionId {
				t.Errorf("TransactionId doesn't match")
			}
			if tran.Description != orig.Description {
				t.Errorf("Description doesn't match")
			}
			if tran.Date != orig.Date {
				t.Errorf("Date doesn't match")
			}

			if len(tran.Splits) != len(orig.Splits) {
				t.Fatalf("Expected %d splits, received %d", len(orig.Splits), len(tran.Splits))
			}

			foundIds := make(map[int64]bool)
			for j := 0; j < len(orig.Splits); j++ {
				origsplit := orig.Splits[j]
				currsplit := curr.Splits[j]

				if tran.Splits[j].TransactionId != tran.TransactionId {
					t.Fatalf("Split TransactionId doesn't match transaction's")
				}

				found := false
				for _, s := range tran.Splits {
					if origsplit.Status == s.Status && origsplit.ImportSplitType == s.ImportSplitType && currsplit.AccountId == s.AccountId && currsplit.SecurityId == s.SecurityId && origsplit.RemoteId == origsplit.RemoteId && origsplit.Number == s.Number && origsplit.Memo == s.Memo && origsplit.Amount == s.Amount {
						if _, ok := foundIds[s.SplitId]; ok {
							continue
						}
						foundIds[s.SplitId] = true
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unable to find matching split: %+v", curr)
				}
			}
		}
	})
}
