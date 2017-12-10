package integration_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestLuaAccounts(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		accounts, err := getAccounts(d.clients[0])
		if err != nil {
			t.Fatalf("Error getting accounts: %s", err)
		}
		accountids := make(Int64Slice, len(*accounts.Accounts))
		for i, s := range *accounts.Accounts {
			accountids[i] = s.AccountId
		}
		accountids.Sort()

		equalityString := ""
		for i := range accountids {
			for j := range accountids {
				if i == j {
					equalityString += "true"
				} else {
					equalityString += "false"
				}
			}
		}

		id := d.accounts[3].AccountId
		simpleLuaTest(t, d.clients[0], []LuaTest{
			{"SecurityId", fmt.Sprintf("return get_accounts()[%d].SecurityId", id), strconv.FormatInt(d.accounts[3].SecurityId, 10)},
			{"Security", fmt.Sprintf("return get_accounts()[%d].Security.SecurityId", id), strconv.FormatInt(d.accounts[3].SecurityId, 10)},
			{"Parent", fmt.Sprintf("return get_accounts()[%d].Parent.AccountId", id), strconv.FormatInt(d.accounts[3].ParentAccountId, 10)},
			{"Name", fmt.Sprintf("return get_accounts()[%d].Name", id), d.accounts[3].Name},
			{"Type", fmt.Sprintf("return get_accounts()[%d].Type", id), strconv.FormatInt(int64(d.accounts[3].Type), 10)},
			{"TypeName", fmt.Sprintf("return get_accounts()[%d].TypeName", id), d.accounts[3].Type.String()},
			{"typename", fmt.Sprintf("return get_accounts()[%d].typename", id), strings.ToLower(d.accounts[3].Type.String())},
			{"Balance()", fmt.Sprintf("return get_accounts()[%d]:Balance().Amount", id), "87.19"},
			{"Balance(1)", fmt.Sprintf("return get_accounts()[%d]:Balance(date.new('2017-10-30')).Amount", id), "5.6"},
			{"Balance(2)", fmt.Sprintf("return get_accounts()[%d]:Balance(date.new(2017, 10, 30), date.new('2017-11-01')).Amount", id), "81.59"},
			{"__tostring", fmt.Sprintf("return get_accounts()[%d]", id), "Expenses/Groceries"},
			{"__eq", `
accounts = get_accounts()
sorted = {}
for id in pairs(accounts) do
	table.insert(sorted, id)
end
str = ""
table.sort(sorted)
for i,idi in ipairs(sorted) do
	for j,idj in ipairs(sorted) do
		if accounts[idi] == accounts[idj] then
			str = str .. "true"
		else
			str = str .. "false"
		end
	end
end
return str`, equalityString},
			{"get_accounts()", `
sorted = {}
for id in pairs(get_accounts()) do
	table.insert(sorted, id)
end
table.sort(sorted)
str = "["
for i,id in ipairs(sorted) do
	str = str .. id .. " "
end
return string.sub(str, 1, -2) .. "]"`, fmt.Sprint(accountids)},
		})
	})
}
