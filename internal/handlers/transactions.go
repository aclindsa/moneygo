package handlers

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

// Split.Status
const (
	Imported   int64 = 1
	Entered          = 2
	Cleared          = 3
	Reconciled       = 4
	Voided           = 5
)

// Split.ImportSplitType
const (
	Default         int64 = 0
	ImportAccount         = 1 // This split belongs to the main account being imported
	SubAccount            = 2 // This split belongs to a sub-account of that being imported
	ExternalAccount       = 3
	TradingAccount        = 4
	Commission            = 5
	Taxes                 = 6
	Fees                  = 7
	Load                  = 8
	IncomeAccount         = 9
	ExpenseAccount        = 10
)

type Split struct {
	SplitId         int64
	TransactionId   int64
	Status          int64
	ImportSplitType int64

	// One of AccountId and SecurityId must be -1
	// In normal splits, AccountId will be valid and SecurityId will be -1. The
	// only case where this is reversed is for transactions that have been
	// imported and not yet associated with an account.
	AccountId  int64
	SecurityId int64

	RemoteId string // unique ID from server, for detecting duplicates
	Number   string // Check or reference number
	Memo     string
	Amount   string // String representation of decimal, suitable for passing to big.Rat.SetString()
}

func GetBigAmount(amt string) (*big.Rat, error) {
	var r big.Rat
	_, success := r.SetString(amt)
	if !success {
		return nil, errors.New("Couldn't convert string amount to big.Rat via SetString()")
	}
	return &r, nil
}

func (s *Split) GetAmount() (*big.Rat, error) {
	return GetBigAmount(s.Amount)
}

func (s *Split) Valid() bool {
	if (s.AccountId == -1) == (s.SecurityId == -1) {
		return false
	}
	_, err := s.GetAmount()
	return err == nil
}

func (s *Split) AlreadyImportedTx(transaction *gorp.Transaction) (bool, error) {
	count, err := transaction.SelectInt("SELECT COUNT(*) from splits where RemoteId=? and AccountId=?", s.RemoteId, s.AccountId)
	return count == 1, err
}

type Transaction struct {
	TransactionId int64
	UserId        int64
	Description   string
	Date          time.Time
	Splits        []*Split `db:"-"`
}

type TransactionList struct {
	Transactions *[]Transaction `json:"transactions"`
}

type AccountTransactionsList struct {
	Account           *Account
	Transactions      *[]Transaction
	TotalTransactions int64
	BeginningBalance  string
	EndingBalance     string
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

// Return a map of security ID's to big.Rat's containing the amount that
// security is imbalanced by
func (t *Transaction) GetImbalancesTx(transaction *gorp.Transaction) (map[int64]big.Rat, error) {
	sums := make(map[int64]big.Rat)

	if !t.Valid() {
		return nil, errors.New("Transaction invalid")
	}

	for i := range t.Splits {
		securityid := t.Splits[i].SecurityId
		if t.Splits[i].AccountId != -1 {
			var err error
			var account *Account
			account, err = GetAccountTx(transaction, t.Splits[i].AccountId, t.UserId)
			if err != nil {
				return nil, err
			}
			securityid = account.SecurityId
		}
		amount, _ := t.Splits[i].GetAmount()
		sum := sums[securityid]
		(&sum).Add(&sum, amount)
		sums[securityid] = sum
	}
	return sums, nil
}

// Returns true if all securities contained in this transaction are balanced,
// false otherwise
func (t *Transaction) Balanced(transaction *gorp.Transaction) (bool, error) {
	var zero big.Rat

	sums, err := t.GetImbalancesTx(transaction)
	if err != nil {
		return false, err
	}

	for _, security_sum := range sums {
		if security_sum.Cmp(&zero) != 0 {
			return false, nil
		}
	}
	return true, nil
}

func GetTransaction(db *DB, transactionid int64, userid int64) (*Transaction, error) {
	var t Transaction

	transaction, err := db.Begin()
	if err != nil {
		return nil, err
	}

	err = transaction.SelectOne(&t, "SELECT * from transactions where UserId=? AND TransactionId=?", userid, transactionid)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	_, err = transaction.Select(&t.Splits, "SELECT * from splits where TransactionId=?", transactionid)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &t, nil
}

func GetTransactions(db *DB, userid int64) (*[]Transaction, error) {
	var transactions []Transaction

	transaction, err := db.Begin()
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

func InsertTransactionTx(transaction *gorp.Transaction, t *Transaction, user *User) error {
	// Map of any accounts with transaction splits being added
	a_map := make(map[int64]bool)
	for i := range t.Splits {
		if t.Splits[i].AccountId != -1 {
			existing, err := transaction.SelectInt("SELECT count(*) from accounts where AccountId=?", t.Splits[i].AccountId)
			if err != nil {
				return err
			}
			if existing != 1 {
				return AccountMissingError{}
			}
			a_map[t.Splits[i].AccountId] = true
		} else if t.Splits[i].SecurityId == -1 {
			return AccountMissingError{}
		}
	}

	//increment versions for all accounts
	var a_ids []int64
	for id := range a_map {
		a_ids = append(a_ids, id)
	}
	// ensure at least one of the splits is associated with an actual account
	if len(a_ids) < 1 {
		return AccountMissingError{}
	}
	err := incrementAccountVersions(transaction, user, a_ids)
	if err != nil {
		return err
	}

	t.UserId = user.UserId
	err = transaction.Insert(t)
	if err != nil {
		return err
	}

	for i := range t.Splits {
		t.Splits[i].TransactionId = t.TransactionId
		t.Splits[i].SplitId = -1
		err = transaction.Insert(t.Splits[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func InsertTransaction(db *DB, t *Transaction, user *User) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	err = InsertTransactionTx(transaction, t, user)
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

func UpdateTransactionTx(transaction *gorp.Transaction, t *Transaction, user *User) error {
	var existing_splits []*Split

	_, err := transaction.Select(&existing_splits, "SELECT * from splits where TransactionId=?", t.TransactionId)
	if err != nil {
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
				return err
			}
			if count != 1 {
				return errors.New("Updated more than one transaction split")
			}
			delete(s_map, t.Splits[i].SplitId)
		} else {
			t.Splits[i].SplitId = -1
			err := transaction.Insert(t.Splits[i])
			if err != nil {
				return err
			}
		}
		if t.Splits[i].AccountId != -1 {
			a_map[t.Splits[i].AccountId] = true
		}
	}

	// Delete any remaining pre-existing splits
	for i := range existing_splits {
		_, ok := s_map[existing_splits[i].SplitId]
		if existing_splits[i].AccountId != -1 {
			a_map[existing_splits[i].AccountId] = true
		}
		if ok {
			_, err := transaction.Delete(existing_splits[i])
			if err != nil {
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
		return err
	}

	count, err := transaction.Update(t)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Updated more than one transaction")
	}

	return nil
}

func DeleteTransaction(db *DB, t *Transaction, user *User) error {
	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	var accountids []int64
	_, err = transaction.Select(&accountids, "SELECT DISTINCT AccountId FROM splits WHERE TransactionId=? AND AccountId != -1", t.TransactionId)
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

func TransactionHandler(w http.ResponseWriter, r *http.Request, db *DB) {
	user, err := GetUserFromSession(db, r)
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

		sqltx, err := db.Begin()
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}

		balanced, err := transaction.Balanced(sqltx)
		if err != nil {
			sqltx.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
		if !transaction.Valid() || !balanced {
			sqltx.Rollback()
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		for i := range transaction.Splits {
			transaction.Splits[i].SplitId = -1
			_, err := GetAccountTx(sqltx, transaction.Splits[i].AccountId, user.UserId)
			if err != nil {
				sqltx.Rollback()
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
		}

		err = InsertTransactionTx(sqltx, &transaction, user)
		if err != nil {
			if _, ok := err.(AccountMissingError); ok {
				WriteError(w, 3 /*Invalid Request*/)
			} else {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
			}
			sqltx.Rollback()
			return
		}

		err = sqltx.Commit()
		if err != nil {
			sqltx.Rollback()
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}

		err = transaction.Write(w)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	} else if r.Method == "GET" {
		transactionid, err := GetURLID(r.URL.Path)

		if err != nil {
			//Return all Transactions
			var al TransactionList
			transactions, err := GetTransactions(db, user.UserId)
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
			transaction, err := GetTransaction(db, transactionid, user.UserId)
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

			sqltx, err := db.Begin()
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			balanced, err := transaction.Balanced(sqltx)
			if err != nil {
				sqltx.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
			if !transaction.Valid() || !balanced {
				sqltx.Rollback()
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			for i := range transaction.Splits {
				_, err := GetAccountTx(sqltx, transaction.Splits[i].AccountId, user.UserId)
				if err != nil {
					sqltx.Rollback()
					WriteError(w, 3 /*Invalid Request*/)
					return
				}
			}

			err = UpdateTransactionTx(sqltx, &transaction, user)
			if err != nil {
				sqltx.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			err = sqltx.Commit()
			if err != nil {
				sqltx.Rollback()
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			err = transaction.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else if r.Method == "DELETE" {
			transactionid, err := GetURLID(r.URL.Path)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			transaction, err := GetTransaction(db, transactionid, user.UserId)
			if err != nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}

			err = DeleteTransaction(db, transaction, user)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}

			WriteSuccess(w)
		}
	}
}

func TransactionsBalanceDifference(transaction *gorp.Transaction, accountid int64, transactions []Transaction) (*big.Rat, error) {
	var pageDifference, tmp big.Rat
	for i := range transactions {
		_, err := transaction.Select(&transactions[i].Splits, "SELECT * FROM splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			return nil, err
		}

		// Sum up the amounts from the splits we're returning so we can return
		// an ending balance
		for j := range transactions[i].Splits {
			if transactions[i].Splits[j].AccountId == accountid {
				rat_amount, err := GetBigAmount(transactions[i].Splits[j].Amount)
				if err != nil {
					return nil, err
				}
				tmp.Add(&pageDifference, rat_amount)
				pageDifference.Set(&tmp)
			}
		}
	}
	return &pageDifference, nil
}

func GetAccountBalance(db *DB, user *User, accountid int64) (*big.Rat, error) {
	var splits []Split
	transaction, err := db.Begin()
	if err != nil {
		return nil, err
	}

	sql := "SELECT DISTINCT splits.* FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=?"
	_, err = transaction.Select(&splits, sql, accountid, user.UserId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	var balance, tmp big.Rat
	for _, s := range splits {
		rat_amount, err := GetBigAmount(s.Amount)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &balance, nil
}

// Assumes accountid is valid and is owned by the current user
func GetAccountBalanceDate(db *DB, user *User, accountid int64, date *time.Time) (*big.Rat, error) {
	var splits []Split
	transaction, err := db.Begin()
	if err != nil {
		return nil, err
	}

	sql := "SELECT DISTINCT splits.* FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=? AND transactions.Date < ?"
	_, err = transaction.Select(&splits, sql, accountid, user.UserId, date)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	var balance, tmp big.Rat
	for _, s := range splits {
		rat_amount, err := GetBigAmount(s.Amount)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &balance, nil
}

func GetAccountBalanceDateRange(db *DB, user *User, accountid int64, begin, end *time.Time) (*big.Rat, error) {
	var splits []Split
	transaction, err := db.Begin()
	if err != nil {
		return nil, err
	}

	sql := "SELECT DISTINCT splits.* FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=? AND transactions.Date >= ? AND transactions.Date < ?"
	_, err = transaction.Select(&splits, sql, accountid, user.UserId, begin, end)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	var balance, tmp big.Rat
	for _, s := range splits {
		rat_amount, err := GetBigAmount(s.Amount)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &balance, nil
}

func GetAccountTransactions(db *DB, user *User, accountid int64, sort string, page uint64, limit uint64) (*AccountTransactionsList, error) {
	var transactions []Transaction
	var atl AccountTransactionsList

	transaction, err := db.Begin()
	if err != nil {
		return nil, err
	}

	var sqlsort, balanceLimitOffset string
	var balanceLimitOffsetArg uint64
	if sort == "date-asc" {
		sqlsort = " ORDER BY transactions.Date ASC"
		balanceLimitOffset = " LIMIT ?"
		balanceLimitOffsetArg = page * limit
	} else if sort == "date-desc" {
		numSplits, err := transaction.SelectInt("SELECT count(*) FROM splits")
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
		sqlsort = " ORDER BY transactions.Date DESC"
		balanceLimitOffset = fmt.Sprintf(" LIMIT %d OFFSET ?", numSplits)
		balanceLimitOffsetArg = (page + 1) * limit
	}

	var sqloffset string
	if page > 0 {
		sqloffset = fmt.Sprintf(" OFFSET %d", page*limit)
	}

	account, err := GetAccountTx(transaction, accountid, user.UserId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	atl.Account = account

	sql := "SELECT DISTINCT transactions.* FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?" + sqlsort + " LIMIT ?" + sqloffset
	_, err = transaction.Select(&transactions, sql, user.UserId, accountid, limit)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	atl.Transactions = &transactions

	pageDifference, err := TransactionsBalanceDifference(transaction, accountid, transactions)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	count, err := transaction.SelectInt("SELECT count(DISTINCT transactions.TransactionId) FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?", user.UserId, accountid)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	atl.TotalTransactions = count

	security, err := GetSecurityTx(transaction, atl.Account.SecurityId, user.UserId)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}
	if security == nil {
		transaction.Rollback()
		return nil, errors.New("Security not found")
	}

	// Sum all the splits for all transaction splits for this account that
	// occurred before the page we're returning
	var amounts []string
	sql = "SELECT splits.Amount FROM splits WHERE splits.AccountId=? AND splits.TransactionId IN (SELECT DISTINCT transactions.TransactionId FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?" + sqlsort + balanceLimitOffset + ")"
	_, err = transaction.Select(&amounts, sql, accountid, user.UserId, accountid, balanceLimitOffsetArg)
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	var tmp, balance big.Rat
	for _, amount := range amounts {
		rat_amount, err := GetBigAmount(amount)
		if err != nil {
			transaction.Rollback()
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}
	atl.BeginningBalance = balance.FloatString(security.Precision)
	atl.EndingBalance = tmp.Add(&balance, pageDifference).FloatString(security.Precision)

	err = transaction.Commit()
	if err != nil {
		transaction.Rollback()
		return nil, err
	}

	return &atl, nil
}

// Return only those transactions which have at least one split pertaining to
// an account
func AccountTransactionsHandler(db *DB, w http.ResponseWriter, r *http.Request,
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

	accountTransactions, err := GetAccountTransactions(db, user, accountid, sort, page, limit)
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
