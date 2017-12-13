package db

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"time"
)

// Price is a mirror of models.Price with the Value broken out into whole and
// fractional components
type Price struct {
	PriceId         int64
	SecurityId      int64
	CurrencyId      int64
	Date            time.Time
	WholeValue      int64
	FractionalValue int64
	RemoteId        string // unique ID from source, for detecting duplicates
}

func NewPrice(p *models.Price) (*Price, error) {
	whole, err := p.Value.Whole()
	if err != nil {
		return nil, err
	}
	fractional, err := p.Value.Fractional(MaxPrecision)
	if err != nil {
		return nil, err
	}
	return &Price{
		PriceId:         p.PriceId,
		SecurityId:      p.SecurityId,
		CurrencyId:      p.CurrencyId,
		Date:            p.Date,
		WholeValue:      whole,
		FractionalValue: fractional,
		RemoteId:        p.RemoteId,
	}, nil
}

func (p Price) Price() *models.Price {
	price := &models.Price{
		PriceId:    p.PriceId,
		SecurityId: p.SecurityId,
		CurrencyId: p.CurrencyId,
		Date:       p.Date,
		RemoteId:   p.RemoteId,
	}
	price.Value.FromParts(p.WholeValue, p.FractionalValue, MaxPrecision)

	return price
}

func (tx *Tx) PriceExists(price *models.Price) (bool, error) {
	p, err := NewPrice(price)
	if err != nil {
		return false, err
	}

	var prices []*Price
	_, err = tx.Select(&prices, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date=? AND WholeValue=? AND FractionalValue=?", p.SecurityId, p.CurrencyId, p.Date, p.WholeValue, p.FractionalValue)
	return len(prices) > 0, err
}

func (tx *Tx) InsertPrice(price *models.Price) error {
	p, err := NewPrice(price)
	if err != nil {
		return err
	}
	err = tx.Insert(p)
	if err != nil {
		return err
	}
	*price = *p.Price()
	return nil
}

func (tx *Tx) GetPrice(priceid, securityid int64) (*models.Price, error) {
	var price Price
	err := tx.SelectOne(&price, "SELECT * from prices where PriceId=? AND SecurityId=?", priceid, securityid)
	if err != nil {
		return nil, err
	}
	return price.Price(), nil
}

func (tx *Tx) GetPrices(securityid int64) (*[]*models.Price, error) {
	var prices []*Price
	var modelprices []*models.Price

	_, err := tx.Select(&prices, "SELECT * from prices where SecurityId=?", securityid)
	if err != nil {
		return nil, err
	}

	for _, p := range prices {
		modelprices = append(modelprices, p.Price())
	}

	return &modelprices, nil
}

// Return the latest price for security in currency units before date
func (tx *Tx) GetLatestPrice(security, currency *models.Security, date *time.Time) (*models.Price, error) {
	var price Price
	err := tx.SelectOne(&price, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date <= ? ORDER BY Date DESC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return price.Price(), nil
}

// Return the earliest price for security in currency units after date
func (tx *Tx) GetEarliestPrice(security, currency *models.Security, date *time.Time) (*models.Price, error) {
	var price Price
	err := tx.SelectOne(&price, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date >= ? ORDER BY Date ASC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return price.Price(), nil
}

func (tx *Tx) UpdatePrice(price *models.Price) error {
	p, err := NewPrice(price)
	if err != nil {
		return err
	}

	count, err := tx.Update(p)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to update 1 price, was going to update %d", count)
	}
	*price = *p.Price()
	return nil
}

func (tx *Tx) DeletePrice(price *models.Price) error {
	p, err := NewPrice(price)
	if err != nil {
		return err
	}

	count, err := tx.Delete(p)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to delete 1 price, was going to delete %d", count)
	}
	*price = *p.Price()
	return nil
}
