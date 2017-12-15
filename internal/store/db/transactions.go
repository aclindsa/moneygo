package db

import (
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"math/big"
	"time"
)

// Split is a mirror of models.Split with the Amount broken out into whole and
// fractional components
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

	// Amount.Whole and Amount.Fractional(MaxPrecision)
	WholeAmount      int64
	FractionalAmount int64
}

func NewSplit(s *models.Split) (*Split, error) {
	whole, err := s.Amount.Whole()
	if err != nil {
		return nil, err
	}
	fractional, err := s.Amount.Fractional(MaxPrecision)
	if err != nil {
		return nil, err
	}
	return &Split{
		SplitId:          s.SplitId,
		TransactionId:    s.TransactionId,
		Status:           s.Status,
		ImportSplitType:  s.ImportSplitType,
		AccountId:        s.AccountId,
		SecurityId:       s.SecurityId,
		RemoteId:         s.RemoteId,
		Number:           s.Number,
		Memo:             s.Memo,
		WholeAmount:      whole,
		FractionalAmount: fractional,
	}, nil
}

func (s Split) Split() *models.Split {
	split := &models.Split{
		SplitId:         s.SplitId,
		TransactionId:   s.TransactionId,
		Status:          s.Status,
		ImportSplitType: s.ImportSplitType,
		AccountId:       s.AccountId,
		SecurityId:      s.SecurityId,
		RemoteId:        s.RemoteId,
		Number:          s.Number,
		Memo:            s.Memo,
	}
	split.Amount.FromParts(s.WholeAmount, s.FractionalAmount, MaxPrecision)

	return split
}

func (tx *Tx) incrementAccountVersions(user *models.User, accountids []int64) error {
	for i := range accountids {
		account, err := tx.GetAccount(accountids[i], user.UserId)
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

func (tx *Tx) InsertTransaction(t *models.Transaction, user *models.User) error {
	// Map of any accounts with transaction splits being added
	a_map := make(map[int64]bool)
	for i := range t.Splits {
		if t.Splits[i].AccountId != -1 {
			existing, err := tx.SelectInt("SELECT count(*) from accounts where AccountId=?", t.Splits[i].AccountId)
			if err != nil {
				return err
			}
			if existing != 1 {
				return store.AccountMissingError{}
			}
			a_map[t.Splits[i].AccountId] = true
		} else if t.Splits[i].SecurityId == -1 {
			return store.AccountMissingError{}
		}
	}

	//increment versions for all accounts
	var a_ids []int64
	for id := range a_map {
		a_ids = append(a_ids, id)
	}
	// ensure at least one of the splits is associated with an actual account
	if len(a_ids) < 1 {
		return store.AccountMissingError{}
	}
	err := tx.incrementAccountVersions(user, a_ids)
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
		s, err := NewSplit(t.Splits[i])
		if err != nil {
			return err
		}
		err = tx.Insert(s)
		if err != nil {
			return err
		}
		*t.Splits[i] = *s.Split()
	}

	return nil
}

func (tx *Tx) SplitExists(s *models.Split) (bool, error) {
	count, err := tx.SelectInt("SELECT COUNT(*) from splits where RemoteId=? and AccountId=?", s.RemoteId, s.AccountId)
	return count == 1, err
}

func (tx *Tx) GetTransaction(transactionid int64, userid int64) (*models.Transaction, error) {
	var t models.Transaction
	var splits []*Split

	err := tx.SelectOne(&t, "SELECT * from transactions where UserId=? AND TransactionId=?", userid, transactionid)
	if err != nil {
		return nil, err
	}

	_, err = tx.Select(&splits, "SELECT * from splits where TransactionId=?", transactionid)
	if err != nil {
		return nil, err
	}

	for _, split := range splits {
		t.Splits = append(t.Splits, split.Split())
	}

	return &t, nil
}

func (tx *Tx) GetTransactions(userid int64) (*[]*models.Transaction, error) {
	var transactions []*models.Transaction

	_, err := tx.Select(&transactions, "SELECT * from transactions where UserId=?", userid)
	if err != nil {
		return nil, err
	}

	for i := range transactions {
		var splits []*Split
		_, err := tx.Select(&splits, "SELECT * from splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			return nil, err
		}
		for _, split := range splits {
			transactions[i].Splits = append(transactions[i].Splits, split.Split())
		}
	}

	return &transactions, nil
}

func (tx *Tx) UpdateTransaction(t *models.Transaction, user *models.User) error {
	var existing_splits []*Split

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
		s, err := NewSplit(t.Splits[i])
		if err != nil {
			return err
		}
		_, ok := s_map[s.SplitId]
		if ok {
			count, err := tx.Update(s)
			if err != nil {
				return err
			}
			if count > 1 {
				return fmt.Errorf("Updated %d transaction splits while attempting to update only 1", count)
			}
			delete(s_map, s.SplitId)
		} else {
			s.SplitId = -1
			err := tx.Insert(s)
			if err != nil {
				return err
			}
		}
		*t.Splits[i] = *s.Split()
		if t.Splits[i].AccountId != -1 {
			a_map[s.AccountId] = true
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
	err = tx.incrementAccountVersions(user, a_ids)
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

func (tx *Tx) DeleteTransaction(t *models.Transaction, user *models.User) error {
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

	err = tx.incrementAccountVersions(user, accountids)
	if err != nil {
		return err
	}

	return nil
}

// Assumes accountid is valid and is owned by the current user
func (tx *Tx) getAccountBalance(xtrasql string, args ...interface{}) (*models.Amount, error) {
	var balance models.Amount

	sql := "FROM splits INNER JOIN transactions ON transactions.TransactionId = splits.TransactionId WHERE splits.AccountId=? AND transactions.UserId=?" + xtrasql
	count, err := tx.SelectInt("SELECT splits.SplitId "+sql+" LIMIT 1", args...)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		type bal struct {
			Whole, Fractional int64
		}
		var b bal
		err := tx.SelectOne(&b, "SELECT sum(splits.WholeAmount) AS Whole, sum(splits.FractionalAmount) AS Fractional "+sql, args...)
		if err != nil {
			return nil, err
		}

		balance.FromParts(b.Whole, b.Fractional, MaxPrecision)
	}

	return &balance, nil
}

func (tx *Tx) GetAccountBalance(user *models.User, accountid int64) (*models.Amount, error) {
	return tx.getAccountBalance("", accountid, user.UserId)
}

func (tx *Tx) GetAccountBalanceDate(user *models.User, accountid int64, date *time.Time) (*models.Amount, error) {
	return tx.getAccountBalance(" AND transactions.Date < ?", accountid, user.UserId, date)
}

func (tx *Tx) GetAccountBalanceDateRange(user *models.User, accountid int64, begin, end *time.Time) (*models.Amount, error) {
	return tx.getAccountBalance(" AND transactions.Date >= ? AND transactions.Date < ?", accountid, user.UserId, begin, end)
}

func (tx *Tx) transactionsBalanceDifference(accountid int64, transactions []*models.Transaction) (*big.Rat, error) {
	var pageDifference big.Rat
	for i := range transactions {
		var splits []*Split
		_, err := tx.Select(&splits, "SELECT * FROM splits where TransactionId=?", transactions[i].TransactionId)
		if err != nil {
			return nil, err
		}

		// Sum up the amounts from the splits we're returning so we can return
		// an ending balance
		for j, s := range splits {
			transactions[i].Splits = append(transactions[i].Splits, s.Split())
			if transactions[i].Splits[j].AccountId == accountid {
				pageDifference.Add(&pageDifference, &transactions[i].Splits[j].Amount.Rat)
			}
		}
	}
	return &pageDifference, nil
}

func (tx *Tx) GetAccountTransactions(user *models.User, accountid int64, sort string, page uint64, limit uint64) (*models.AccountTransactionsList, error) {
	var transactions []*models.Transaction
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

	account, err := tx.GetAccount(accountid, user.UserId)
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

	pageDifference, err := tx.transactionsBalanceDifference(accountid, transactions)
	if err != nil {
		return nil, err
	}

	count, err := tx.SelectInt("SELECT count(DISTINCT transactions.TransactionId) FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?", user.UserId, accountid)
	if err != nil {
		return nil, err
	}
	atl.TotalTransactions = count

	security, err := tx.GetSecurity(atl.Account.SecurityId, user.UserId)
	if err != nil {
		return nil, err
	}
	if security == nil {
		return nil, errors.New("Security not found")
	}

	// Sum all the splits for all transaction splits for this account that
	// occurred before the page we're returning
	sql = "FROM splits AS s INNER JOIN (SELECT DISTINCT transactions.Date, transactions.TransactionId FROM transactions INNER JOIN splits ON transactions.TransactionId = splits.TransactionId WHERE transactions.UserId=? AND splits.AccountId=?" + sqlsort + balanceLimitOffset + ") as t ON s.TransactionId = t.TransactionId WHERE s.AccountId=?"
	count, err = tx.SelectInt("SELECT count(*) "+sql, user.UserId, accountid, balanceLimitOffsetArg, accountid)
	if err != nil {
		return nil, err
	}

	var balance models.Amount

	// Don't attempt to 'sum()' the splits if none exist, because it is
	// supposed to return null/nil in this case, which makes gorp angry since
	// we're using SelectInt()
	if count > 0 {
		whole, err := tx.SelectInt("SELECT sum(s.WholeAmount) "+sql, user.UserId, accountid, balanceLimitOffsetArg, accountid)
		if err != nil {
			return nil, err
		}
		fractional, err := tx.SelectInt("SELECT sum(s.FractionalAmount) "+sql, user.UserId, accountid, balanceLimitOffsetArg, accountid)
		if err != nil {
			return nil, err
		}
		balance.FromParts(whole, fractional, MaxPrecision)
	}

	atl.BeginningBalance = balance
	atl.EndingBalance.Rat.Add(&balance.Rat, pageDifference)

	return &atl, nil
}
