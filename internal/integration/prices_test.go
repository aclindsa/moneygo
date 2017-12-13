package integration_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func createPrice(client *http.Client, price *models.Price) (*models.Price, error) {
	var p models.Price
	err := create(client, price, &p, "/v1/securities/"+strconv.FormatInt(price.SecurityId, 10)+"/prices/")
	return &p, err
}

func getPrice(client *http.Client, priceid, securityid int64) (*models.Price, error) {
	var p models.Price
	err := read(client, &p, "/v1/securities/"+strconv.FormatInt(securityid, 10)+"/prices/"+strconv.FormatInt(priceid, 10))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func getPrices(client *http.Client, securityid int64) (*models.PriceList, error) {
	var pl models.PriceList
	err := read(client, &pl, "/v1/securities/"+strconv.FormatInt(securityid, 10)+"/prices/")
	if err != nil {
		return nil, err
	}
	return &pl, nil
}

func updatePrice(client *http.Client, price *models.Price) (*models.Price, error) {
	var p models.Price
	err := update(client, price, &p, "/v1/securities/"+strconv.FormatInt(price.SecurityId, 10)+"/prices/"+strconv.FormatInt(price.PriceId, 10))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func deletePrice(client *http.Client, p *models.Price) error {
	err := remove(client, "/v1/securities/"+strconv.FormatInt(p.SecurityId, 10)+"/prices/"+strconv.FormatInt(p.PriceId, 10))
	if err != nil {
		return err
	}
	return nil
}

func TestCreatePrice(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].prices); i++ {
			orig := data[0].prices[i]
			p := d.prices[i]

			if p.PriceId == 0 {
				t.Errorf("Unable to create price: %+v", p)
			}
			if p.SecurityId != d.securities[orig.SecurityId].SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if p.CurrencyId != d.securities[orig.CurrencyId].SecurityId {
				t.Errorf("CurrencyId doesn't match")
			}
			if !p.Date.Equal(orig.Date) {
				t.Errorf("Date doesn't match")
			}
			if p.Value.Cmp(&orig.Value.Rat) != 0 {
				t.Errorf("Value doesn't match")
			}
			if p.RemoteId != orig.RemoteId {
				t.Errorf("RemoteId doesn't match")
			}
		}
	})
}

func TestGetPrice(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].prices); i++ {
			orig := data[0].prices[i]
			curr := d.prices[i]

			userid := data[0].securities[orig.SecurityId].UserId
			p, err := getPrice(d.clients[userid], curr.PriceId, curr.SecurityId)
			if err != nil {
				t.Fatalf("Error fetching price: %s\n", err)
			}
			if p.SecurityId != d.securities[orig.SecurityId].SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if p.CurrencyId != d.securities[orig.CurrencyId].SecurityId {
				t.Errorf("CurrencyId doesn't match")
			}
			if !p.Date.Equal(orig.Date) {
				t.Errorf("Date doesn't match")
			}
			if p.Value.Cmp(&orig.Value.Rat) != 0 {
				t.Errorf("Value doesn't match")
			}
			if p.RemoteId != orig.RemoteId {
				t.Errorf("RemoteId doesn't match")
			}
		}
	})
}

func TestGetPrices(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for origsecurityid, security := range d.securities {
			if data[0].securities[origsecurityid].UserId != 0 {
				continue
			}

			pl, err := getPrices(d.clients[0], security.SecurityId)
			if err != nil {
				t.Fatalf("Error fetching prices: %s\n", err)
			}

			numprices := 0
			foundIds := make(map[int64]bool)
			for i := 0; i < len(data[0].prices); i++ {
				orig := data[0].prices[i]

				if orig.SecurityId != int64(origsecurityid) {
					continue
				}
				numprices += 1

				found := false
				for _, p := range *pl.Prices {
					if p.SecurityId == d.securities[orig.SecurityId].SecurityId && p.CurrencyId == d.securities[orig.CurrencyId].SecurityId && p.Date.Equal(orig.Date) && p.Value.Cmp(&orig.Value.Rat) == 0 && p.RemoteId == orig.RemoteId {
						if _, ok := foundIds[p.PriceId]; ok {
							continue
						}
						foundIds[p.PriceId] = true
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unable to find matching price: %+v", orig)
				}
			}

			if pl.Prices == nil {
				if numprices != 0 {
					t.Fatalf("Expected %d prices, received 0", numprices)
				}
			} else if numprices != len(*pl.Prices) {
				t.Fatalf("Expected %d prices, received %d", numprices, len(*pl.Prices))
			}
		}
	})
}

func TestUpdatePrice(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].prices); i++ {
			orig := data[0].prices[i]
			curr := d.prices[i]

			tmp := curr.SecurityId
			curr.SecurityId = curr.CurrencyId
			curr.CurrencyId = tmp
			curr.Value = NewAmount("5.55")
			curr.Date = time.Date(2019, time.June, 5, 12, 5, 6, 7, time.UTC)
			curr.RemoteId = "something"

			userid := data[0].securities[orig.SecurityId].UserId
			p, err := updatePrice(d.clients[userid], &curr)
			if err != nil {
				t.Fatalf("Error updating price: %s\n", err)
			}

			if p.SecurityId != curr.SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if p.CurrencyId != curr.CurrencyId {
				t.Errorf("CurrencyId doesn't match")
			}
			if !p.Date.Equal(curr.Date) {
				t.Errorf("Date doesn't match")
			}
			if p.Value.Cmp(&curr.Value.Rat) != 0 {
				t.Errorf("Value doesn't match")
			}
			if p.RemoteId != curr.RemoteId {
				t.Errorf("RemoteId doesn't match")
			}
		}
	})
}

func TestDeletePrice(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].prices); i++ {
			orig := data[0].prices[i]
			curr := d.prices[i]

			userid := data[0].securities[orig.SecurityId].UserId
			err := deletePrice(d.clients[userid], &curr)
			if err != nil {
				t.Fatalf("Error deleting price: %s\n", err)
			}

			_, err = getPrice(d.clients[userid], curr.PriceId, curr.SecurityId)
			if err == nil {
				t.Fatalf("Expected error fetching deleted price")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error fetching deleted price: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error fetching deleted price")
			}
		}
	})
}
