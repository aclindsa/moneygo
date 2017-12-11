package handlers

import (
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"log"
	"net/http"
)

func CreatePriceIfNotExist(tx store.Tx, price *models.Price) error {
	if len(price.RemoteId) == 0 {
		// Always create a new price if we can't match on the RemoteId
		err := tx.InsertPrice(price)
		if err != nil {
			return err
		}
		return nil
	}

	exists, err := tx.PriceExists(price)
	if err != nil {
		return err
	}
	if exists {
		return nil // price already exists
	}

	err = tx.InsertPrice(price)
	if err != nil {
		return err
	}
	return nil
}

func PriceHandler(r *http.Request, context *Context, user *models.User, securityid int64) ResponseWriterWriter {
	security, err := context.Tx.GetSecurity(securityid, user.UserId)
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
		_, err = context.Tx.GetSecurity(price.CurrencyId, user.UserId)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		err = context.Tx.InsertPrice(&price)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		return ResponseWrapper{201, &price}
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all this security's prices
			var pl models.PriceList

			prices, err := context.Tx.GetPrices(security.SecurityId)
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

		price, err := context.Tx.GetPrice(priceid, security.SecurityId)
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

			_, err = context.Tx.GetSecurity(price.SecurityId, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}
			_, err = context.Tx.GetSecurity(price.CurrencyId, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.UpdatePrice(&price)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &price
		} else if r.Method == "DELETE" {
			price, err := context.Tx.GetPrice(priceid, security.SecurityId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.DeletePrice(price)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
