package integration_test

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"testing"
)

type LuaTest struct {
	Name     string
	Lua      string
	Expected string
}

func simpleLuaTest(t *testing.T, client *http.Client, tests []LuaTest) {
	t.Helper()
	for _, lt := range tests {
		lua := fmt.Sprintf(`function test()
	%s
end

function generate()
    t = tabulation.new(0)
    t:title(tostring(test()))
	return t
end`, lt.Lua)
		r := models.Report{
			Name: lt.Name,
			Lua:  lua,
		}
		report, err := createReport(client, &r)
		if err != nil {
			t.Fatalf("Error creating report: %s", err)
		}

		tab, err := tabulateReport(client, report.ReportId)
		if err != nil {
			t.Fatalf("Error tabulating report: %s", err)
		}

		if tab.Title != lt.Expected {
			t.Errorf("%s: Returned '%s', expected '%s'", lt.Name, tab.Title, lt.Expected)
		}
	}
}
