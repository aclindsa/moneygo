package handlers_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func createPrice(client *http.Client, price *handlers.Price) (*handlers.Price, error) {
	var p handlers.Price
	err := create(client, price, &p, "/v1/prices/")
	return &p, err
}

func getPrice(client *http.Client, priceid int64) (*handlers.Price, error) {
	var p handlers.Price
	err := read(client, &p, "/v1/prices/"+strconv.FormatInt(priceid, 10))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func getPrices(client *http.Client) (*handlers.PriceList, error) {
	var pl handlers.PriceList
	err := read(client, &pl, "/v1/prices/")
	if err != nil {
		return nil, err
	}
	return &pl, nil
}

func updatePrice(client *http.Client, price *handlers.Price) (*handlers.Price, error) {
	var p handlers.Price
	err := update(client, price, &p, "/v1/prices/"+strconv.FormatInt(price.PriceId, 10))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func deletePrice(client *http.Client, p *handlers.Price) error {
	err := remove(client, "/v1/prices/"+strconv.FormatInt(p.PriceId, 10))
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
			if p.Date != orig.Date {
				t.Errorf("Date doesn't match")
			}
			if p.Value != orig.Value {
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
			p, err := getPrice(d.clients[userid], curr.PriceId)
			if err != nil {
				t.Fatalf("Error fetching price: %s\n", err)
			}
			if p.SecurityId != d.securities[orig.SecurityId].SecurityId {
				t.Errorf("SecurityId doesn't match")
			}
			if p.CurrencyId != d.securities[orig.CurrencyId].SecurityId {
				t.Errorf("CurrencyId doesn't match")
			}
			if p.Date != orig.Date {
				t.Errorf("Date doesn't match")
			}
			if p.Value != orig.Value {
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
		pl, err := getPrices(d.clients[0])
		if err != nil {
			t.Fatalf("Error fetching prices: %s\n", err)
		}

		numprices := 0
		foundIds := make(map[int64]bool)
		for i := 0; i < len(data[0].prices); i++ {
			orig := data[0].prices[i]

			if data[0].securities[orig.SecurityId].UserId != 0 {
				continue
			}
			numprices += 1

			found := false
			for _, p := range *pl.Prices {
				if p.SecurityId == d.securities[orig.SecurityId].SecurityId && p.CurrencyId == d.securities[orig.CurrencyId].SecurityId && p.Date != orig.Date && p.Value != orig.Value && p.RemoteId != orig.RemoteId {
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

		if numprices != len(*pl.Prices) {
			t.Fatalf("Expected %d prices, received %d", numprices, len(*pl.Prices))
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
			curr.Value = "5.55"
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
			if p.Date != curr.Date {
				t.Errorf("Date doesn't match")
			}
			if p.Value != curr.Value {
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

			_, err = getPrice(d.clients[userid], curr.PriceId)
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
