package handlers

import (
	"gopkg.in/gorp.v1"
	"time"
)

type Price struct {
	PriceId    int64
	SecurityId int64
	CurrencyId int64
	Date       time.Time
	Value      string // String representation of decimal price of Security in Currency units, suitable for passing to big.Rat.SetString()
	RemoteId   string // unique ID from source, for detecting duplicates
}

func InsertPriceTx(transaction *gorp.Transaction, p *Price) error {
	err := transaction.Insert(p)
	if err != nil {
		return err
	}
	return nil
}

func CreatePriceIfNotExist(transaction *gorp.Transaction, price *Price) error {
	if len(price.RemoteId) == 0 {
		// Always create a new price if we can't match on the RemoteId
		err := InsertPriceTx(transaction, price)
		if err != nil {
			return err
		}
		return nil
	}

	var prices []*Price

	_, err := transaction.Select(&prices, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date=? AND Value=?", price.SecurityId, price.CurrencyId, price.Date, price.Value)
	if err != nil {
		return err
	}

	if len(prices) > 0 {
		return nil // price already exists
	}

	err = InsertPriceTx(transaction, price)
	if err != nil {
		return err
	}
	return nil
}

// Return the latest price for security in currency units before date
func GetLatestPrice(transaction *gorp.Transaction, security, currency *Security, date *time.Time) (*Price, error) {
	var p Price
	err := transaction.SelectOne(&p, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date <= ? ORDER BY Date DESC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Return the earliest price for security in currency units after date
func GetEarliestPrice(transaction *gorp.Transaction, security, currency *Security, date *time.Time) (*Price, error) {
	var p Price
	err := transaction.SelectOne(&p, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date >= ? ORDER BY Date ASC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Return the price for security in currency closest to date
func GetClosestPriceTx(transaction *gorp.Transaction, security, currency *Security, date *time.Time) (*Price, error) {
	earliest, _ := GetEarliestPrice(transaction, security, currency, date)
	latest, err := GetLatestPrice(transaction, security, currency, date)

	// Return early if either earliest or latest are invalid
	if earliest == nil {
		return latest, err
	} else if err != nil {
		return earliest, nil
	}

	howlate := earliest.Date.Sub(*date)
	howearly := date.Sub(latest.Date)
	if howearly < howlate {
		return latest, nil
	} else {
		return earliest, nil
	}
}

func GetClosestPrice(tx *Tx, security, currency *Security, date *time.Time) (*Price, error) {
	return GetClosestPriceTx(tx, security, currency, date)
}
