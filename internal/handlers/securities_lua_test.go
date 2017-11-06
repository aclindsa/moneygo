package handlers_test

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/handlers"
	"strconv"
	"testing"
)

type LuaTest struct {
	Name     string
	Lua      string
	Expected string
}

func TestLuaSecurities(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		defaultSecurity, err := getSecurity(d.clients[0], d.users[0].DefaultCurrency)
		if err != nil {
			t.Fatalf("Error getting default security: %s", err)
		}

		for _, lt := range []LuaTest{
			{"SecurityId", `return get_default_currency().SecurityId`, strconv.FormatInt(defaultSecurity.SecurityId, 10)},
			{"Name", `return get_default_currency().Name`, defaultSecurity.Name},
			{"Description", `return get_default_currency().Description`, defaultSecurity.Description},
			{"Symbol", `return get_default_currency().Symbol`, defaultSecurity.Symbol},
			{"Precision", `return get_default_currency().Precision`, strconv.FormatInt(int64(defaultSecurity.Precision), 10)},
			{"Type", `return get_default_currency().Type`, strconv.FormatInt(int64(defaultSecurity.Type), 10)},
			{"AlternateId", `return get_default_currency().AlternateId`, defaultSecurity.AlternateId},
		} {
			lua := fmt.Sprintf(`function test()
	%s
end

function generate()
    t = tabulation.new(0)
    t:title(tostring(test()))
	return t
end`, lt.Lua)
			r := handlers.Report{
				Name: lt.Name,
				Lua:  lua,
			}
			report, err := createReport(d.clients[0], &r)
			if err != nil {
				t.Fatalf("Error creating report: %s", err)
			}

			tab, err := tabulateReport(d.clients[0], report.ReportId)
			if err != nil {
				t.Fatalf("Error tabulating report: %s", err)
			}

			if tab.Title != lt.Expected {
				t.Errorf("%s: Returned '%s', expected '%s'", lt.Name, tab.Title, lt.Expected)
			}
		}
	})
}
