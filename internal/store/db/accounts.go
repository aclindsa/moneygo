package db

import (
	"errors"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
)

func (tx *Tx) GetAccount(accountid int64, userid int64) (*models.Account, error) {
	var account models.Account

	err := tx.SelectOne(&account, "SELECT * from accounts where UserId=? AND AccountId=?", userid, accountid)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (tx *Tx) GetAccounts(userid int64) (*[]*models.Account, error) {
	var accounts []*models.Account

	_, err := tx.Select(&accounts, "SELECT * from accounts where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

func (tx *Tx) FindMatchingAccounts(account *models.Account) (*[]*models.Account, error) {
	var accounts []*models.Account

	_, err := tx.Select(&accounts, "SELECT * from accounts where UserId=? AND SecurityId=? AND Type=? AND Name=? AND ParentAccountId=? ORDER BY AccountId ASC", account.UserId, account.SecurityId, account.Type, account.Name, account.ParentAccountId)
	if err != nil {
		return nil, err
	}
	return &accounts, nil
}

func (tx *Tx) insertUpdateAccount(account *models.Account, insert bool) error {
	found := make(map[int64]bool)
	if !insert {
		found[account.AccountId] = true
	}
	parentid := account.ParentAccountId
	depth := 0
	for parentid != -1 {
		depth += 1
		if depth > 100 {
			return store.TooMuchNestingError{}
		}

		var a models.Account
		err := tx.SelectOne(&a, "SELECT * from accounts where AccountId=?", parentid)
		if err != nil {
			return store.ParentAccountMissingError{}
		}

		// Insertion by itself can never result in circular dependencies
		if insert {
			break
		}

		found[parentid] = true
		parentid = a.ParentAccountId
		if _, ok := found[parentid]; ok {
			return store.CircularAccountsError{}
		}
	}

	if insert {
		err := tx.Insert(account)
		if err != nil {
			return err
		}
	} else {
		oldacct, err := tx.GetAccount(account.AccountId, account.UserId)
		if err != nil {
			return err
		}

		account.AccountVersion = oldacct.AccountVersion + 1

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

func (tx *Tx) InsertAccount(account *models.Account) error {
	return tx.insertUpdateAccount(account, true)
}

func (tx *Tx) UpdateAccount(account *models.Account) error {
	return tx.insertUpdateAccount(account, false)
}

func (tx *Tx) DeleteAccount(account *models.Account) error {
	if account.ParentAccountId != -1 {
		// Re-parent splits to this account's parent account if this account isn't a root account
		_, err := tx.Exec("UPDATE splits SET AccountId=? WHERE AccountId=?", account.ParentAccountId, account.AccountId)
		if err != nil {
			return err
		}
	} else {
		// Delete splits if this account is a root account
		_, err := tx.Exec("DELETE FROM splits WHERE AccountId=?", account.AccountId)
		if err != nil {
			return err
		}
	}

	// Re-parent child accounts to this account's parent account
	_, err := tx.Exec("UPDATE accounts SET ParentAccountId=? WHERE ParentAccountId=?", account.ParentAccountId, account.AccountId)
	if err != nil {
		return err
	}

	count, err := tx.Delete(account)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Was going to delete more than one account")
	}

	return nil
}
