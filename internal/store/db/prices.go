package db

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"time"
)

func (tx *Tx) PriceExists(price *models.Price) (bool, error) {
	var prices []*models.Price
	_, err := tx.Select(&prices, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date=? AND Value=?", price.SecurityId, price.CurrencyId, price.Date, price.Value)
	return len(prices) > 0, err
}

func (tx *Tx) InsertPrice(price *models.Price) error {
	return tx.Insert(price)
}

func (tx *Tx) GetPrice(priceid, securityid int64) (*models.Price, error) {
	var price models.Price
	err := tx.SelectOne(&price, "SELECT * from prices where PriceId=? AND SecurityId=?", priceid, securityid)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

func (tx *Tx) GetPrices(securityid int64) (*[]*models.Price, error) {
	var prices []*models.Price

	_, err := tx.Select(&prices, "SELECT * from prices where SecurityId=?", securityid)
	if err != nil {
		return nil, err
	}
	return &prices, nil
}

// Return the latest price for security in currency units before date
func (tx *Tx) GetLatestPrice(security, currency *models.Security, date *time.Time) (*models.Price, error) {
	var price models.Price
	err := tx.SelectOne(&price, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date <= ? ORDER BY Date DESC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

// Return the earliest price for security in currency units after date
func (tx *Tx) GetEarliestPrice(security, currency *models.Security, date *time.Time) (*models.Price, error) {
	var price models.Price
	err := tx.SelectOne(&price, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date >= ? ORDER BY Date ASC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

func (tx *Tx) UpdatePrice(price *models.Price) error {
	count, err := tx.Update(price)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to update 1 price, was going to update %d", count)
	}
	return nil
}

func (tx *Tx) DeletePrice(price *models.Price) error {
	count, err := tx.Delete(price)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to delete 1 price, was going to delete %d", count)
	}
	return nil
}
