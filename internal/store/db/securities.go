package db

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
)

// MaxPrexision denotes the maximum valid value for models.Security.Precision.
// This constant is used when storing amounts in securities into the database,
// so it must not be changed without appropriately migrating the database.
const MaxPrecision uint64 = 15

func init() {
	if MaxPrecision < models.MaxPrecision {
		panic("db.MaxPrecision must be >= models.MaxPrecision")
	}
}

func (tx *Tx) GetSecurity(securityid int64, userid int64) (*models.Security, error) {
	var s models.Security

	err := tx.SelectOne(&s, "SELECT * from securities where UserId=? AND SecurityId=?", userid, securityid)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (tx *Tx) GetSecurities(userid int64) (*[]*models.Security, error) {
	var securities []*models.Security

	_, err := tx.Select(&securities, "SELECT * from securities where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &securities, nil
}

func (tx *Tx) FindMatchingSecurities(security *models.Security) (*[]*models.Security, error) {
	var securities []*models.Security

	_, err := tx.Select(&securities, "SELECT * from securities where UserId=? AND Type=? AND AlternateId=? AND Preciseness=?", security.UserId, security.Type, security.AlternateId, security.Precision)
	if err != nil {
		return nil, err
	}
	return &securities, nil
}

func (tx *Tx) InsertSecurity(s *models.Security) error {
	err := tx.Insert(s)
	if err != nil {
		return err
	}
	return nil
}

func (tx *Tx) UpdateSecurity(security *models.Security) error {
	count, err := tx.Update(security)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to update 1 security, was going to update %d", count)
	}
	return nil
}

func (tx *Tx) DeleteSecurity(s *models.Security) error {
	// First, ensure no accounts are using this security
	accounts, err := tx.SelectInt("SELECT count(*) from accounts where UserId=? and SecurityId=?", s.UserId, s.SecurityId)

	if accounts != 0 {
		return store.SecurityInUseError{"One or more accounts still use this security"}
	}

	user, err := tx.GetUser(s.UserId)
	if err != nil {
		return err
	} else if user.DefaultCurrency == s.SecurityId {
		return store.SecurityInUseError{"Cannot delete security which is user's default currency"}
	}

	// Remove all prices involving this security (either of this security, or
	// using it as a currency)
	_, err = tx.Exec("DELETE FROM prices WHERE SecurityId=? OR CurrencyId=?", s.SecurityId, s.SecurityId)
	if err != nil {
		return err
	}

	count, err := tx.Delete(s)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to delete 1 security, was going to delete %d", count)
	}
	return nil
}
