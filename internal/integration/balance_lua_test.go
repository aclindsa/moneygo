package integration_test

import (
	"fmt"
	"testing"
)

func TestLuaBalances(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		accountid := d.accounts[3].AccountId
		symbol := d.securities[data[0].accounts[3].SecurityId].Symbol

		simpleLuaTest(t, d.clients[0], []LuaTest{
			{"Account:Balance()", fmt.Sprintf("return get_accounts()[%d]:Balance()", accountid), symbol + " 87.19"},
			{"Account:Balance(1)", fmt.Sprintf("return get_accounts()[%d]:Balance(date.new('2017-10-30')).Amount", accountid), "5.6"},
			{"Account:Balance(2)", fmt.Sprintf("return get_accounts()[%d]:Balance(date.new(2017, 10, 30), date.new('2017-11-01')).Amount", accountid), "81.59"},
			{"Security", fmt.Sprintf("return get_accounts()[%d]:Balance().Security.Symbol", accountid), symbol},
			{"__eq", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 10, 30)) == (act:Balance(date.new('2017-10-29')) + 0.0)", accountid), "true"},
			{"not __eq", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 10, 30)) == act:Balance(date.new('2017-11-01'))", accountid), "false"},
			{"__lt", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 10, 14)) < act:Balance(date.new('2017-10-16'))", accountid), "true"},
			{"not __lt", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 11, 01)) < act:Balance(date.new('2017-10-16'))", accountid), "false"},
			{"__le", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 10, 14)) <= act:Balance(date.new('2017-10-16'))", accountid), "true"},
			{"__le (=)", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 10, 16)) <= act:Balance(date.new('2017-10-17'))", accountid), "true"},
			{"not __le", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new(2017, 11, 01)) <= act:Balance(date.new('2017-10-16'))", accountid), "false"},
			{"__add", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) + act:Balance(date.new(2017, 10, 30), date.new('2017-11-01'))", accountid), symbol + " 87.19"},
			{"__add number", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) + 9", accountid), symbol + " 14.60"},
			{"__add to number", fmt.Sprintf("act = get_accounts()[%d]; return 5.489 + act:Balance(date.new(2017, 10, 30), date.new('2017-11-01'))", accountid), symbol + " 87.08"},
			{"__sub", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) - act:Balance(date.new(2017, 10, 30), date.new('2017-11-01'))", accountid), symbol + " -75.99"},
			{"__sub number", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) - 5", accountid), symbol + " 0.60"},
			{"__sub from number", fmt.Sprintf("act = get_accounts()[%d]; return 100 - act:Balance(date.new(2017, 10, 30), date.new('2017-11-01'))", accountid), symbol + " 18.41"},
			{"__mul", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) * act:Balance(date.new(2017, 10, 30), date.new('2017-11-01'))", accountid), symbol + " 456.90"},
			{"__mul number", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) * 5", accountid), symbol + " 28.00"},
			{"__mul with number", fmt.Sprintf("act = get_accounts()[%d]; return 11.1111 * act:Balance(date.new('2017-10-30')) * 5", accountid), symbol + " 311.11"},
			{"__div", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) / act:Balance(date.new(2017, 10, 30), date.new('2017-11-01'))", accountid), symbol + " 0.07"},
			{"__div number", fmt.Sprintf("act = get_accounts()[%d]; return act:Balance(date.new('2017-10-30')) / 5", accountid), symbol + " 1.12"},
			{"__div with number", fmt.Sprintf("act = get_accounts()[%d]; return 11.1111 / act:Balance(date.new('2017-10-30'))", accountid), symbol + " 1.98"},
			{"__unm", fmt.Sprintf("act = get_accounts()[%d]; return -act:Balance(date.new('2017-10-30'))", accountid), symbol + " -5.60"},
		})
	})
}
