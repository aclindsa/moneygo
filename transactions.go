package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/gorp.v1"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Split struct {
	SplitId       int64
	TransactionId int64
	AccountId     int64
	Number        string // Check or reference number
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

const (
	Entered    int64 = 1
	Cleared                      = 2
	Reconciled                   = 3
	Voided                       = 4
)

type Transaction struct {
	TransactionId int64
	UserId        int64
	Description   string
	Status        int64
	Date          time.Time
	Splits        []*Split `db:"-"`
}

type TransactionList struct {
	Transactions *[]Transaction `json:"transactions"`
}

type AccountTransactionsList struct {
	Account      *Account       `json:"account"`
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

func (atl *AccountTransactionsList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(atl)
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

	err = transaction.SelectOne(&t, "SELECT * from transactions where UserId=? AND TransactionId=?", userid, transactionid)
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

func incrementAccountVersions(transaction *gorp.Transaction, user *User, accountids []int64) error {
	for i := range accountids {
		account, err := GetAccountTx(transaction, accountids[i], user.UserId)
		if err != nil {
			return err
		}
		account.AccountVersion++
		count, err := transaction.Update(account)
		if err != nil {
			return err
		}
		if count != 1 {
			return errors.New("Updated more than one account")
		}
	}
	return nil
}

type AccountMissingError struct{}

func (ame AccountMissingError) Error() string {
	return "Account missing"
}

func InsertTransaction(t *Transaction, user *User) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	// Map of any accounts with transaction splits being added
	a_map := make(map[int64]bool)
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
		a_map[t.Splits[i].AccountId] = true
	}

	//increment versions for all accounts
	var a_ids []int64
	for id := range a_map {
		a_ids = append(a_ids, id)
	}
	err = incrementAccountVersions(transaction, user, a_ids)
	if err != nil {
		transaction.Rollback()
		return err
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

func UpdateTransaction(t *Transaction, user *User) error {
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

	// Map of any accounts with transaction splits being added
	a_map := make(map[int64]bool)

	// Make a map with any existing splits for this transaction
	s_map := make(map[int64]bool)
	for i := range existing_splits {
		s_map[existing_splits[i].SplitId] = true
	}

	// Insert splits, updating any pre-existing ones
	for i := range t.Splits {
		t.Splits[i].TransactionId = t.TransactionId
		_, ok := s_map[t.Splits[i].SplitId]
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
		a_map[t.Splits[i].AccountId] = true
	}

	// Delete any remaining pre-existing splits
	for i := range existing_splits {
		_, ok := s_map[existing_splits[i].SplitId]
		a_map[existing_splits[i].AccountId] = true
		if ok {
			_, err := transaction.Delete(existing_splits[i])
			if err != nil {
				transaction.Rollback()
				return err
			}
		}
	}

	// Increment versions for all accounts with modified splits
	var a_ids []int64
	for id := range a_map {
		a_ids = append(a_ids, id)
	}
	err = incrementAccountVersions(transaction, user, a_ids)
	if err != nil {
		transaction.Rollback()
		return err
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

func DeleteTransaction(t *Transaction, user *User) error {
	transaction, err := DB.Begin()
	if err != nil {
		return err
	}

	var accountids []int64
	_, err = transaction.Select(&accountids, "SELECT DISTINCT AccountId FROM splits WHERE TransactionId=?", t.TransactionId)
	if err != nil {
		transaction.Rollback()
		return err
	}

	_, err = transaction.Exec("DELETE FROM splits WHERE TransactionId=?", t.TransactionId)
	if err != nil {
		transaction.Rollback()
		return err
	}

	count, err := transaction.Delete(t)
	if err != nil {
		transaction.Rollback()
		return err
	}
	if count != 1 {
		transaction.Rollback()
		return errors.New("Deleted more than one transaction")
	}

	err = incrementAccountVersions(transaction, user, accountids)
	if err != nil {
		transaction.Rollback()
		return err
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

		err = InsertTransaction(&transaction, user)
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

			err = UpdateTransaction(&transaction, user)
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

			err = DeleteTransaction(transaction, user)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}

func GetAccountTransactions(user *User, accountid int64, sort string, page uint64, limit uint64) (*AccountTransactionsList, error) {
	var transactions []Transaction
	var atl AccountTransactionsList

	var sqlsort string
	if sort == "date-asc" {
		sqlsort = " ORDER BY transactions.Date ASC"
	} else if sort == "date-desc" {
		sqlsort = " ORDER BY transactions.Date DESC"
	}

	var sqloffset string
	if page > 0 {
		sqloffset = fmt.Sprintf(" OFFSET %d", page*limit)
	}

	transaction, err := DB.Begin()
	if err != nil {
		return nil, err
	}

	account, err := GetAccountTx(transaction, accountid, user.UserId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	atl.Account = account

	sql := "SELECT transactions.* from transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?" + sqlsort + " LIMIT ?" + sqloffset
	_, err = transaction.Select(&transactions, sql, user.UserId, accountid, limit)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	atl.Transactions = &transactions

	for i := range transactions {
		_, err = transaction.Select(&transactions[i].Splits, "SELECT * from splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &atl, nil
}

// Return only those transactions which have at least one split pertaining to
// an account
func AccountTransactionsHandler(w http.ResponseWriter, r *http.Request,
	user *User, accountid int64) {

	var page uint64 = 0
	var limit uint64 = 50
	var sort string = "date-desc"

	query, _ := url.ParseQuery(r.URL.RawQuery)

	pagestring := query.Get("page")
	if pagestring != "" {
		p, err := strconv.ParseUint(pagestring, 10, 0)
		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		page = p
	}

	limitstring := query.Get("limit")
	if limitstring != "" {
		l, err := strconv.ParseUint(limitstring, 10, 0)
		if err != nil || l > 100 {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		limit = l
	}

	sortstring := query.Get("sort")
	if sortstring != "" {
		if sortstring != "date-asc" && sortstring != "date-desc" {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}
		sort = sortstring
	}

	accountTransactions, err := GetAccountTransactions(user, accountid, sort, page, limit)
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}

	err = accountTransactions.Write(w)
	if err != nil {
		WriteError(w, 999 /*Internal Error*/)
		log.Print(err)
		return
	}
}
