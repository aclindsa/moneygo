package integration_test

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
)

// Int64Slice attaches the methods of int64 to []int64, sorting in increasing order.
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p Int64Slice) Sort() { sort.Sort(p) }

func TestLuaSecurities(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		defaultSecurity, err := getSecurity(d.clients[0], d.users[0].DefaultCurrency)
		if err != nil {
			t.Fatalf("Error getting default security: %s", err)
		}
		securities, err := getSecurities(d.clients[0])
		if err != nil {
			t.Fatalf("Error getting securities: %s", err)
		}
		securityids := make(Int64Slice, len(*securities.Securities))
		for i, s := range *securities.Securities {
			securityids[i] = s.SecurityId
		}
		securityids.Sort()

		equalityString := ""
		for i := range securityids {
			for j := range securityids {
				if i == j {
					equalityString += "true"
				} else {
					equalityString += "false"
				}
			}
		}

		simpleLuaTest(t, d.clients[0], []LuaTest{
			{"__tostring", `return get_default_currency()`, fmt.Sprintf("%s - %s (%s)", defaultSecurity.Name, defaultSecurity.Description, defaultSecurity.Symbol)},
			{"SecurityId", `return get_default_currency().SecurityId`, strconv.FormatInt(defaultSecurity.SecurityId, 10)},
			{"Name", `return get_default_currency().Name`, defaultSecurity.Name},
			{"Description", `return get_default_currency().Description`, defaultSecurity.Description},
			{"Symbol", `return get_default_currency().Symbol`, defaultSecurity.Symbol},
			{"Precision", `return get_default_currency().Precision`, strconv.FormatInt(int64(defaultSecurity.Precision), 10)},
			{"Type", `return get_default_currency().Type`, strconv.FormatInt(int64(defaultSecurity.Type), 10)},
			{"AlternateId", `return get_default_currency().AlternateId`, defaultSecurity.AlternateId},
			{"__eq", `
securities = get_securities()
sorted = {}
for id in pairs(securities) do
	table.insert(sorted, id)
end
str = ""
table.sort(sorted)
for i,idi in ipairs(sorted) do
	for j,idj in ipairs(sorted) do
		if securities[idi] == securities[idj] then
			str = str .. "true"
		else
			str = str .. "false"
		end
	end
end
return str`, equalityString},
			{"get_securities()", `
sorted = {}
for id in pairs(get_securities()) do
	table.insert(sorted, id)
end
table.sort(sorted)
str = "["
for i,id in ipairs(sorted) do
	str = str .. id .. " "
end
return string.sub(str, 1, -2) .. "]"`, fmt.Sprint(securityids)},
		})
	})
}
