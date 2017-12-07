package handlers

import (
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store/db"
	"log"
	"net/http"
	"time"
)

func CreatePriceIfNotExist(tx *db.Tx, price *models.Price) error {
	if len(price.RemoteId) == 0 {
		// Always create a new price if we can't match on the RemoteId
		err := tx.Insert(price)
		if err != nil {
			return err
		}
		return nil
	}

	var prices []*models.Price

	_, err := tx.Select(&prices, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date=? AND Value=?", price.SecurityId, price.CurrencyId, price.Date, price.Value)
	if err != nil {
		return err
	}

	if len(prices) > 0 {
		return nil // price already exists
	}

	err = tx.Insert(price)
	if err != nil {
		return err
	}
	return nil
}

func GetPrice(tx *db.Tx, priceid, securityid int64) (*models.Price, error) {
	var p models.Price
	err := tx.SelectOne(&p, "SELECT * from prices where PriceId=? AND SecurityId=?", priceid, securityid)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetPrices(tx *db.Tx, securityid int64) (*[]*models.Price, error) {
	var prices []*models.Price

	_, err := tx.Select(&prices, "SELECT * from prices where SecurityId=?", securityid)
	if err != nil {
		return nil, err
	}
	return &prices, nil
}

// Return the latest price for security in currency units before date
func GetLatestPrice(tx *db.Tx, security, currency *models.Security, date *time.Time) (*models.Price, error) {
	var p models.Price
	err := tx.SelectOne(&p, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date <= ? ORDER BY Date DESC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Return the earliest price for security in currency units after date
func GetEarliestPrice(tx *db.Tx, security, currency *models.Security, date *time.Time) (*models.Price, error) {
	var p models.Price
	err := tx.SelectOne(&p, "SELECT * from prices where SecurityId=? AND CurrencyId=? AND Date >= ? ORDER BY Date ASC LIMIT 1", security.SecurityId, currency.SecurityId, date)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Return the price for security in currency closest to date
func GetClosestPrice(tx *db.Tx, security, currency *models.Security, date *time.Time) (*models.Price, error) {
	earliest, _ := GetEarliestPrice(tx, security, currency, date)
	latest, err := GetLatestPrice(tx, security, currency, date)

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

func PriceHandler(r *http.Request, context *Context, user *models.User, securityid int64) ResponseWriterWriter {
	security, err := GetSecurity(context.Tx, securityid, user.UserId)
	if err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	if r.Method == "POST" {
		var price models.Price
		if err := ReadJSON(r, &price); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		price.PriceId = -1

		if price.SecurityId != security.SecurityId {
			return NewError(3 /*Invalid Request*/)
		}
		_, err = GetSecurity(context.Tx, price.CurrencyId, user.UserId)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		err = context.Tx.Insert(&price)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		return ResponseWrapper{201, &price}
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all this security's prices
			var pl models.PriceList

			prices, err := GetPrices(context.Tx, security.SecurityId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			pl.Prices = prices
			return &pl
		}

		priceid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		price, err := GetPrice(context.Tx, priceid, security.SecurityId)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		return price
	} else {
		priceid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		if r.Method == "PUT" {
			var price models.Price
			if err := ReadJSON(r, &price); err != nil || price.PriceId != priceid {
				return NewError(3 /*Invalid Request*/)
			}

			_, err = GetSecurity(context.Tx, price.SecurityId, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			_, err = GetSecurity(context.Tx, price.CurrencyId, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			count, err := context.Tx.Update(&price)
			if err != nil || count != 1 {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &price
		} else if r.Method == "DELETE" {
			price, err := GetPrice(context.Tx, priceid, security.SecurityId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			count, err := context.Tx.Delete(price)
			if err != nil || count != 1 {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
