package handlers

import (
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func SplitAlreadyImported(tx *Tx, s *models.Split) (bool, error) {
	count, err := tx.SelectInt("SELECT COUNT(*) from splits where RemoteId=? and AccountId=?", s.RemoteId, s.AccountId)
	return count == 1, err
}

// Return a map of security ID's to big.Rat's containing the amount that
// security is imbalanced by
func GetTransactionImbalances(tx *Tx, t *models.Transaction) (map[int64]big.Rat, error) {
	sums := make(map[int64]big.Rat)

	if !t.Valid() {
		return nil, errors.New("Transaction invalid")
	}

	for i := range t.Splits {
		securityid := t.Splits[i].SecurityId
		if t.Splits[i].AccountId != -1 {
			var err error
			var account *models.Account
			account, err = GetAccount(tx, t.Splits[i].AccountId, t.UserId)
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
func TransactionBalanced(tx *Tx, t *models.Transaction) (bool, error) {
	var zero big.Rat

	sums, err := GetTransactionImbalances(tx, t)
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

func GetTransaction(tx *Tx, transactionid int64, userid int64) (*models.Transaction, error) {
	var t models.Transaction

	err := tx.SelectOne(&t, "SELECT * from transactions where UserId=? AND TransactionId=?", userid, transactionid)
	if err != nil {
		return nil, err
	}

	_, err = tx.Select(&t.Splits, "SELECT * from splits where TransactionId=?", transactionid)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func GetTransactions(tx *Tx, userid int64) (*[]models.Transaction, error) {
	var transactions []models.Transaction

	_, err := tx.Select(&transactions, "SELECT * from transactions where UserId=?", userid)
	if err != nil {
		return nil, err
	}

	for i := range transactions {
		_, err := tx.Select(&transactions[i].Splits, "SELECT * from splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			return nil, err
		}
	}

	return &transactions, nil
}

func incrementAccountVersions(tx *Tx, user *models.User, accountids []int64) error {
	for i := range accountids {
		account, err := GetAccount(tx, accountids[i], user.UserId)
		if err != nil {
			return err
		}
		account.AccountVersion++
		count, err := tx.Update(account)
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

func InsertTransaction(tx *Tx, t *models.Transaction, user *models.User) error {
	// Map of any accounts with transaction splits being added
	a_map := make(map[int64]bool)
	for i := range t.Splits {
		if t.Splits[i].AccountId != -1 {
			existing, err := tx.SelectInt("SELECT count(*) from accounts where AccountId=?", t.Splits[i].AccountId)
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
	err := incrementAccountVersions(tx, user, a_ids)
	if err != nil {
		return err
	}

	t.UserId = user.UserId
	err = tx.Insert(t)
	if err != nil {
		return err
	}

	for i := range t.Splits {
		t.Splits[i].TransactionId = t.TransactionId
		t.Splits[i].SplitId = -1
		err = tx.Insert(t.Splits[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateTransaction(tx *Tx, t *models.Transaction, user *models.User) error {
	var existing_splits []*models.Split

	_, err := tx.Select(&existing_splits, "SELECT * from splits where TransactionId=?", t.TransactionId)
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
			count, err := tx.Update(t.Splits[i])
			if err != nil {
				return err
			}
			if count > 1 {
				return fmt.Errorf("Updated %d transaction splits while attempting to update only 1", count)
			}
			delete(s_map, t.Splits[i].SplitId)
		} else {
			t.Splits[i].SplitId = -1
			err := tx.Insert(t.Splits[i])
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
			_, err := tx.Delete(existing_splits[i])
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
	err = incrementAccountVersions(tx, user, a_ids)
	if err != nil {
		return err
	}

	count, err := tx.Update(t)
	if err != nil {
		return err
	}
	if count > 1 {
		return fmt.Errorf("Updated %d transactions (expected 1)", count)
	}

	return nil
}

func DeleteTransaction(tx *Tx, t *models.Transaction, user *models.User) error {
	var accountids []int64
	_, err := tx.Select(&accountids, "SELECT DISTINCT AccountId FROM splits WHERE TransactionId=? AND AccountId != -1", t.TransactionId)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM splits WHERE TransactionId=?", t.TransactionId)
	if err != nil {
		return err
	}

	count, err := tx.Delete(t)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Deleted more than one transaction")
	}

	err = incrementAccountVersions(tx, user, accountids)
	if err != nil {
		return err
	}

	return nil
}

func TransactionHandler(r *http.Request, context *Context) ResponseWriterWriter {
	user, err := GetUserFromSession(context.Tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method == "POST" {
		var transaction models.Transaction
		if err := ReadJSON(r, &transaction); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		transaction.TransactionId = -1
		transaction.UserId = user.UserId

		if len(transaction.Splits) == 0 {
			return NewError(3 /*Invalid Request*/)
		}

		for i := range transaction.Splits {
			transaction.Splits[i].SplitId = -1
			_, err := GetAccount(context.Tx, transaction.Splits[i].AccountId, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
		}

		balanced, err := TransactionBalanced(context.Tx, &transaction)
		if err != nil {
			return NewError(999 /*Internal Error*/)
		}
		if !transaction.Valid() || !balanced {
			return NewError(3 /*Invalid Request*/)
		}

		err = InsertTransaction(context.Tx, &transaction, user)
		if err != nil {
			if _, ok := err.(AccountMissingError); ok {
				return NewError(3 /*Invalid Request*/)
			} else {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
		}

		return &transaction
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all Transactions
			var al models.TransactionList
			transactions, err := GetTransactions(context.Tx, user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			al.Transactions = transactions
			return &al
		} else {
			//Return Transaction with this Id
			transactionid, err := context.NextID()
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			transaction, err := GetTransaction(context.Tx, transactionid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			return transaction
		}
	} else {
		transactionid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		if r.Method == "PUT" {
			var transaction models.Transaction
			if err := ReadJSON(r, &transaction); err != nil || transaction.TransactionId != transactionid {
				return NewError(3 /*Invalid Request*/)
			}
			transaction.UserId = user.UserId

			balanced, err := TransactionBalanced(context.Tx, &transaction)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			if !transaction.Valid() || !balanced {
				return NewError(3 /*Invalid Request*/)
			}

			if len(transaction.Splits) == 0 {
				return NewError(3 /*Invalid Request*/)
			}

			for i := range transaction.Splits {
				_, err := GetAccount(context.Tx, transaction.Splits[i].AccountId, user.UserId)
				if err != nil {
					return NewError(3 /*Invalid Request*/)
				}
			}

			err = UpdateTransaction(context.Tx, &transaction, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &transaction
		} else if r.Method == "DELETE" {
			transaction, err := GetTransaction(context.Tx, transactionid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = DeleteTransaction(context.Tx, transaction, user)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}

func TransactionsBalanceDifference(tx *Tx, accountid int64, transactions []models.Transaction) (*big.Rat, error) {
	var pageDifference, tmp big.Rat
	for i := range transactions {
		_, err := tx.Select(&transactions[i].Splits, "SELECT * FROM splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			return nil, err
		}

		// Sum up the amounts from the splits we're returning so we can return
		// an ending balance
		for j := range transactions[i].Splits {
			if transactions[i].Splits[j].AccountId == accountid {
				rat_amount, err := models.GetBigAmount(transactions[i].Splits[j].Amount)
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

func GetAccountBalance(tx *Tx, user *models.User, accountid int64) (*big.Rat, error) {
	var splits []models.Split

	sql := "SELECT DISTINCT splits.* FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=?"
	_, err := tx.Select(&splits, sql, accountid, user.UserId)
	if err != nil {
		return nil, err
	}

	var balance, tmp big.Rat
	for _, s := range splits {
		rat_amount, err := models.GetBigAmount(s.Amount)
		if err != nil {
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}

	return &balance, nil
}

// Assumes accountid is valid and is owned by the current user
func GetAccountBalanceDate(tx *Tx, user *models.User, accountid int64, date *time.Time) (*big.Rat, error) {
	var splits []models.Split

	sql := "SELECT DISTINCT splits.* FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=? AND transactions.Date < ?"
	_, err := tx.Select(&splits, sql, accountid, user.UserId, date)
	if err != nil {
		return nil, err
	}

	var balance, tmp big.Rat
	for _, s := range splits {
		rat_amount, err := models.GetBigAmount(s.Amount)
		if err != nil {
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}

	return &balance, nil
}

func GetAccountBalanceDateRange(tx *Tx, user *models.User, accountid int64, begin, end *time.Time) (*big.Rat, error) {
	var splits []models.Split

	sql := "SELECT DISTINCT splits.* FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=? AND transactions.Date >= ? AND transactions.Date < ?"
	_, err := tx.Select(&splits, sql, accountid, user.UserId, begin, end)
	if err != nil {
		return nil, err
	}

	var balance, tmp big.Rat
	for _, s := range splits {
		rat_amount, err := models.GetBigAmount(s.Amount)
		if err != nil {
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}

	return &balance, nil
}

func GetAccountTransactions(tx *Tx, user *models.User, accountid int64, sort string, page uint64, limit uint64) (*models.AccountTransactionsList, error) {
	var transactions []models.Transaction
	var atl models.AccountTransactionsList

	var sqlsort, balanceLimitOffset string
	var balanceLimitOffsetArg uint64
	if sort == "date-asc" {
		sqlsort = " ORDER BY transactions.Date ASC, transactions.TransactionId ASC"
		balanceLimitOffset = " LIMIT ?"
		balanceLimitOffsetArg = page * limit
	} else if sort == "date-desc" {
		numSplits, err := tx.SelectInt("SELECT count(*) FROM splits")
		if err != nil {
			return nil, err
		}
		sqlsort = " ORDER BY transactions.Date DESC, transactions.TransactionId DESC"
		balanceLimitOffset = fmt.Sprintf(" LIMIT %d OFFSET ?", numSplits)
		balanceLimitOffsetArg = (page + 1) * limit
	}

	var sqloffset string
	if page > 0 {
		sqloffset = fmt.Sprintf(" OFFSET %d", page*limit)
	}

	account, err := GetAccount(tx, accountid, user.UserId)
	if err != nil {
		return nil, err
	}
	atl.Account = account

	sql := "SELECT DISTINCT transactions.* FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?" + sqlsort + " LIMIT ?" + sqloffset
	_, err = tx.Select(&transactions, sql, user.UserId, accountid, limit)
	if err != nil {
		return nil, err
	}
	atl.Transactions = &transactions

	pageDifference, err := TransactionsBalanceDifference(tx, accountid, transactions)
	if err != nil {
		return nil, err
	}

	count, err := tx.SelectInt("SELECT count(DISTINCT transactions.TransactionId) FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?", user.UserId, accountid)
	if err != nil {
		return nil, err
	}
	atl.TotalTransactions = count

	security, err := GetSecurity(tx, atl.Account.SecurityId, user.UserId)
	if err != nil {
		return nil, err
	}
	if security == nil {
		return nil, errors.New("Security not found")
	}

	// Sum all the splits for all transaction splits for this account that
	// occurred before the page we're returning
	var amounts []string
	sql = "SELECT s.Amount FROM splits AS s INNER JOIN (SELECT DISTINCT transactions.Date, transactions.TransactionId FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?" + sqlsort + balanceLimitOffset + ") as t ON s.TransactionId = t.TransactionId WHERE s.AccountId=?"
	_, err = tx.Select(&amounts, sql, user.UserId, accountid, balanceLimitOffsetArg, accountid)
	if err != nil {
		return nil, err
	}

	var tmp, balance big.Rat
	for _, amount := range amounts {
		rat_amount, err := models.GetBigAmount(amount)
		if err != nil {
			return nil, err
		}
		tmp.Add(&balance, rat_amount)
		balance.Set(&tmp)
	}
	atl.BeginningBalance = balance.FloatString(security.Precision)
	atl.EndingBalance = tmp.Add(&balance, pageDifference).FloatString(security.Precision)

	return &atl, nil
}

// Return only those transactions which have at least one split pertaining to
// an account
func AccountTransactionsHandler(context *Context, r *http.Request, user *models.User, accountid int64) ResponseWriterWriter {
	var page uint64 = 0
	var limit uint64 = 50
	var sort string = "date-desc"

	query, _ := url.ParseQuery(r.URL.RawQuery)

	pagestring := query.Get("page")
	if pagestring != "" {
		p, err := strconv.ParseUint(pagestring, 10, 0)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		page = p
	}

	limitstring := query.Get("limit")
	if limitstring != "" {
		l, err := strconv.ParseUint(limitstring, 10, 0)
		if err != nil || l > 100 {
			return NewError(3 /*Invalid Request*/)
		}
		limit = l
	}

	sortstring := query.Get("sort")
	if sortstring != "" {
		if sortstring != "date-asc" && sortstring != "date-desc" {
			return NewError(3 /*Invalid Request*/)
		}
		sort = sortstring
	}

	accountTransactions, err := GetAccountTransactions(context.Tx, user, accountid, sort, page, limit)
	if err != nil {
		log.Print(err)
		return NewError(999 /*Internal Error*/)
	}

	return accountTransactions
}
