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
		}
	})
}
