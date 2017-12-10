package db

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
)

func (tx *Tx) UsernameExists(username string) (bool, error) {
	existing, err := tx.SelectInt("SELECT count(*) from users where Username=?", username)
	return existing != 0, err
}

func (tx *Tx) InsertUser(user *models.User) error {
	return tx.Insert(user)
}

func (tx *Tx) GetUser(userid int64) (*models.User, error) {
	var u models.User

	err := tx.SelectOne(&u, "SELECT * from users where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (tx *Tx) GetUserByUsername(username string) (*models.User, error) {
	var u models.User

	err := tx.SelectOne(&u, "SELECT * from users where Username=?", username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (tx *Tx) UpdateUser(user *models.User) error {
	count, err := tx.Update(user)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to update 1 user, was going to update %d", count)
	}
	return nil
}

func (tx *Tx) DeleteUser(user *models.User) error {
	count, err := tx.Delete(user)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to delete 1 user, was going to delete %d", count)
	}
	_, err = tx.Exec("DELETE FROM prices WHERE prices.SecurityId IN (SELECT securities.SecurityId FROM securities WHERE securities.UserId=?)", user.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM splits WHERE splits.TransactionId IN (SELECT transactions.TransactionId FROM transactions WHERE transactions.UserId=?)", user.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM transactions WHERE transactions.UserId=?", user.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM securities WHERE securities.UserId=?", user.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM accounts WHERE accounts.UserId=?", user.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM reports WHERE reports.UserId=?", user.UserId)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM sessions WHERE sessions.UserId=?", user.UserId)
	if err != nil {
		return err
	}

	return nil
}
