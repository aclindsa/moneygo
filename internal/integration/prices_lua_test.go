package integration_test

import (
	"fmt"
	"strconv"
	"testing"
)

func TestLuaPrices(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		security := d.securities[1]
		currency := d.securities[0]

		simpleLuaTest(t, d.clients[0], []LuaTest{
			{"Security:ClosestPrice", fmt.Sprintf("secs = get_securities(); return secs[%d]:ClosestPrice(secs[%d], date.new('2016-11-19'))", security.SecurityId, currency.SecurityId), fmt.Sprintf("225.24 %s (%s)", currency.Symbol, security.Symbol)},
			{"Security:ClosestPrice(2)", fmt.Sprintf("secs = get_securities(); return secs[%d]:ClosestPrice(secs[%d], date.new('2017-01-04'))", security.SecurityId, currency.SecurityId), fmt.Sprintf("226.58 %s (%s)", currency.Symbol, security.Symbol)},
			{"PriceId", fmt.Sprintf("secs = get_securities(); return secs[%d]:ClosestPrice(secs[%d], date.new('2016-11-19')).PriceId", security.SecurityId, currency.SecurityId), strconv.FormatInt(d.prices[0].PriceId, 10)},
			{"Security", fmt.Sprintf("secs = get_securities(); return secs[%d]:ClosestPrice(secs[%d], date.new('2016-11-19')).Security == secs[%d]", security.SecurityId, currency.SecurityId, security.SecurityId), "true"},
			{"Currency", fmt.Sprintf("secs = get_securities(); return secs[%d]:ClosestPrice(secs[%d], date.new('2016-11-19')).Currency == secs[%d]", security.SecurityId, currency.SecurityId, currency.SecurityId), "true"},
			{"Value", fmt.Sprintf("secs = get_securities(); return secs[%d]:ClosestPrice(secs[%d], date.new('2098-11-09')).Value", security.SecurityId, currency.SecurityId), "227.21"},
		})
	})
}
