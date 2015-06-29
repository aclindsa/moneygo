package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type Split struct {
	SplitId       int64
	TransactionId int64
	AccountId     int64
	Number        int64 // Check or reference number
	Memo          string
	Amount        string // String representation of decimal, suitable for passing to big.Rat.SetString()
	Debit         bool
}

func (s *Split) GetAmount() (*big.Rat, error) {
	var r big.Rat
	_, success := r.SetString(s.Amount)
	if !success {
		return nil, errors.New("Couldn't convert Split.Amount to big.Rat via SetString()")
	}
	return &r, nil
}

func (s *Split) Valid() bool {
	_, err := s.GetAmount()
	return err == nil
}

type TransactionStatus int64

const (
	Entered    TransactionStatus = 1
	Cleared                      = 2
	Reconciled                   = 3
	Voided                       = 4
)

type Transaction struct {
	TransactionId int64
	UserId        int64
	Description   string
	Status        TransactionStatus
	Date          time.Time
	Splits        []*Split `db:"-"`
}

type TransactionList struct {
	Transactions *[]Transaction `json:"transactions"`
}

func (t *Transaction) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(t)
}

func (t *Transaction) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(t)
}

func (tl *TransactionList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(tl)
}

func (t *Transaction) Valid() bool {
	for i := range t.Splits {
		if !t.Splits[i].Valid() {
			return false
		}
	}
	return true
}

func (t *Transaction) Balanced() bool {
	var zero, sum big.Rat
	if !t.Valid() {
		return false // TODO Open question: should we report an error here instead?
	}
	for i := range t.Splits {
		amount, _ := t.Splits[i].GetAmount()
		if t.Splits[i].Debit {
			sum.Add(&sum, amount)
		} else {
			sum.Sub(&sum, amount)
		}
	}
	return sum.Cmp(&zero) == 0
}

func GetTransaction(transactionid int64, userid int64) (*Transaction, error) {
	var t Transaction

	transaction, err := DB.Begin()
	if err != nil {
		return nil, err
	}

	err = transaction.SelectOne(&t, "SELECT * from transaction where UserId=? AND TransactionId=?", userid, transactionid)
	if err != nil {
		return nil, err
	}

	_, err = transaction.Select(&t.Splits, "SELECT * from splits where TransactionId=?", transactionid)
	if err != nil {
		return nil, err
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &t, nil
}

func GetTransactions(userid int64) (*[]Transaction, error) {
	var transactions []Transaction

	transaction, err := DB.Begin()
	if err != nil {
		return nil, err
	}

	_, err = transaction.Select(&transactions, "SELECT * from transactions where UserId=?", userid)
	if err != nil {
		return nil, err
	}

	for i := range transactions {
		_, err := transaction.Select(&transactions[i].Splits, "SELECT * from splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			return nil, err
		}
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &transactions, nil
}

type AccountMissingError struct{}

func (ame AccountMissingError) Error() string {
	return "Account missing"
}

func InsertTransaction(t *Transaction) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	for i := range t.Splits {
		existing, err := transaction.SelectInt("SELECT count(*) from accounts where AccountId=?", t.Splits[i].AccountId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if existing != 1 {
			transaction.Rollback()
			return AccountMissingError{}
		}
	}

	err = transaction.Insert(t)
	if err != nil {
		transaction.Rollback()
		return err
	}

	for i := range t.Splits {
		t.Splits[i].TransactionId = t.TransactionId
		t.Splits[i].SplitId = -1
		err = transaction.Insert(t.Splits[i])
		if err != nil {
			transaction.Rollback()
			return err
		}
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func UpdateTransaction(t *Transaction) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	var existing_splits []*Split

	_, err = transaction.Select(&existing_splits, "SELECT * from splits where TransactionId=?", t.TransactionId)
	if err != nil {
		transaction.Rollback()
		return err
	}

	// Make a map with any existing splits for this transaction
	m := make(map[int64]int64)
	for i := range existing_splits {
		m[existing_splits[i].SplitId] = existing_splits[i].SplitId
	}

	// Insert splits, updating any pre-existing ones
	for i := range t.Splits {
		t.Splits[i].TransactionId = t.TransactionId
		_, ok := m[t.Splits[i].SplitId]
		if ok {
			count, err := transaction.Update(t.Splits[i])
			if err != nil {
				transaction.Rollback()
				return err
			}
			if count != 1 {
				transaction.Rollback()
				return errors.New("Updated more than one transaction split")
			}
		} else {
			t.Splits[i].SplitId = -1
			err := transaction.Insert(t.Splits[i])
			if err != nil {
				transaction.Rollback()
				return err
			}
		}
	}

	// Delete any remaining pre-existing splits
	for i := range existing_splits {
		s, ok := m[existing_splits[i].SplitId]
		if ok {
			_, err := transaction.Delete(s)
			if err != nil {
				transaction.Rollback()
				return err
			}
		}
	}

	count, err := transaction.Update(t)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if count != 1 {
		transaction.Rollback()
		return errors.New("Updated more than one transaction")
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		WriteError(w, 1 /*Not Signed In*/)
		return
	}

	if r.Method == "POST" {
		transaction_json := r.PostFormValue("transaction")
		if transaction_json == "" {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		var transaction Transaction
		err := transaction.Read(transaction_json)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		transaction.TransactionId = -1
		transaction.UserId = user.UserId

		if !transaction.Valid() || !transaction.Balanced() {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		for i := range transaction.Splits {
			transaction.Splits[i].SplitId = -1
			_, err := GetAccount(transaction.Splits[i].AccountId, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
		}

		err = InsertTransaction(&transaction)
		if err != nil {
			if _, ok := err.(AccountMissingError); ok {
				WriteError(w, 3 /*Invalid Request*/)
			} else {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
			}
			return
		}

		WriteSuccess(w)
	} else if r.Method == "GET" {
		transactionid, err := GetURLID(r.URL.Path)
		if err != nil {
			//Return all Transactions
			var al TransactionList
			transactions, err := GetTransactions(user.UserId)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
			al.Transactions = transactions
			err = (&al).Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else {
			//Return Transaction with this Id
			transaction, err := GetTransaction(transactionid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			err = transaction.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		}
	} else {
		transactionid, err := GetURLID(r.URL.Path)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		if r.Method == "PUT" {
			transaction_json := r.PostFormValue("transaction")
			if transaction_json == "" {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			var transaction Transaction
			err := transaction.Read(transaction_json)
			if err != nil || transaction.TransactionId != transactionid {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			transaction.UserId = user.UserId

			if !transaction.Valid() || !transaction.Balanced() {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			for i := range transaction.Splits {
				transaction.Splits[i].SplitId = -1
				_, err := GetAccount(transaction.Splits[i].AccountId, user.UserId)
				if err != nil {
					WriteError(w, 3 /*Invalid Request*/)
					return
				}
			}

			err = UpdateTransaction(&transaction)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		} else if r.Method == "DELETE" {
			transactionid, err := GetURLID(r.URL.Path)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			transaction, err := GetTransaction(transactionid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			count, err := DB.Delete(&transaction)
			if count != 1 || err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}
