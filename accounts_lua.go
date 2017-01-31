package main

import (
	"context"
	"errors"
	"github.com/yuin/gopher-lua"
	"math/big"
)

const luaAccountTypeName = "account"

func luaContextGetAccounts(L *lua.LState) (map[int64]*Account, error) {
	var account_map map[int64]*Account

	ctx := L.Context()

	account_map, ok := ctx.Value(accountsContextKey).(map[int64]*Account)
	if !ok {
		user, ok := ctx.Value(userContextKey).(*User)
		if !ok {
			return nil, errors.New("Couldn't find User in lua's Context")
		}

		accounts, err := GetAccounts(user.UserId)
		if err != nil {
			return nil, err
		}

		account_map = make(map[int64]*Account)
		for i := range *accounts {
			account_map[(*accounts)[i].AccountId] = &(*accounts)[i]
		}

		ctx = context.WithValue(ctx, accountsContextKey, account_map)
		L.SetContext(ctx)
	}

	return account_map, nil
}

func luaGetAccounts(L *lua.LState) int {
	account_map, err := luaContextGetAccounts(L)
	if err != nil {
		panic("luaGetAccounts couldn't fetch accounts")
	}

	table := L.NewTable()

	for accountid := range account_map {
		table.RawSetInt(int(accountid), AccountToLua(L, account_map[accountid]))
	}

	L.Push(table)
	return 1
}

// Registers my account type to given L.
func luaRegisterAccounts(L *lua.LState) {
	mt := L.NewTypeMetatable(luaAccountTypeName)
	L.SetGlobal("account", mt)
	L.SetField(mt, "__index", L.NewFunction(luaAccount__index))
	L.SetField(mt, "__tostring", L.NewFunction(luaAccount__tostring))
	L.SetField(mt, "__eq", L.NewFunction(luaAccount__eq))
	L.SetField(mt, "__metatable", lua.LString("protected"))

	getAccountsFn := L.NewFunction(luaGetAccounts)
	L.SetField(mt, "get_all", getAccountsFn)
	// also register the get_accounts function as a global in its own right
	L.SetGlobal("get_accounts", getAccountsFn)
}

func AccountToLua(L *lua.LState, account *Account) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = account
	L.SetMetatable(ud, L.GetTypeMetatable(luaAccountTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Account and returns this *Account.
func luaCheckAccount(L *lua.LState, n int) *Account {
	ud := L.CheckUserData(n)
	if account, ok := ud.Value.(*Account); ok {
		return account
	}
	L.ArgError(n, "account expected")
	return nil
}

func luaAccount__index(L *lua.LState) int {
	a := luaCheckAccount(L, 1)
	field := L.CheckString(2)

	switch field {
	case "AccountId", "accountid":
		L.Push(lua.LNumber(float64(a.AccountId)))
	case "Security", "security":
		security_map, err := luaContextGetSecurities(L)
		if err != nil {
			panic("account.security couldn't fetch securities")
		}
		if security, ok := security_map[a.SecurityId]; ok {
			L.Push(SecurityToLua(L, security))
		} else {
			panic("SecurityId not in lua security_map")
		}
	case "Parent", "parent", "ParentAccount", "parentaccount":
		if a.ParentAccountId == -1 {
			L.Push(lua.LNil)
		} else {
			account_map, err := luaContextGetAccounts(L)
			if err != nil {
				panic("account.parent couldn't fetch accounts")
			}
			if parent, ok := account_map[a.ParentAccountId]; ok {
				L.Push(AccountToLua(L, parent))
			} else {
				panic("ParentAccountId not in lua account_map")
			}
		}
	case "Name", "name":
		L.Push(lua.LString(a.Name))
	case "Type", "type":
		L.Push(lua.LNumber(float64(a.Type)))
	case "Balance", "balance":
		L.Push(L.NewFunction(luaAccountBalance))
	default:
		L.ArgError(2, "unexpected account attribute: "+field)
	}

	return 1
}

func luaAccountBalance(L *lua.LState) int {
	a := luaCheckAccount(L, 1)

	ctx := L.Context()
	user, ok := ctx.Value(userContextKey).(*User)
	if !ok {
		panic("Couldn't find User in lua's Context")
	}
	security_map, err := luaContextGetSecurities(L)
	if err != nil {
		panic("account.security couldn't fetch securities")
	}
	security, ok := security_map[a.SecurityId]
	if !ok {
		panic("SecurityId not in lua security_map")
	}
	date := luaWeakCheckTime(L, 2)
	var b Balance
	var rat *big.Rat
	if date != nil {
		end := luaWeakCheckTime(L, 3)
		if end != nil {
			rat, err = GetAccountBalanceDateRange(user, a.AccountId, date, end)
		} else {
			rat, err = GetAccountBalanceDate(user, a.AccountId, date)
		}
	} else {
		rat, err = GetAccountBalance(user, a.AccountId)
	}
	if err != nil {
		panic("Failed to GetAccountBalance:" + err.Error())
	}
	b.Amount = rat
	b.Security = security
	L.Push(BalanceToLua(L, &b))

	return 1
}

func luaAccount__tostring(L *lua.LState) int {
	a := luaCheckAccount(L, 1)

	account_map, err := luaContextGetAccounts(L)
	if err != nil {
		panic("luaGetAccounts couldn't fetch accounts")
	}

	full_name := a.Name
	for a.ParentAccountId != -1 {
		a = account_map[a.ParentAccountId]
		full_name = a.Name + "/" + full_name
	}

	L.Push(lua.LString(full_name))

	return 1
}

func luaAccount__eq(L *lua.LState) int {
	a := luaCheckAccount(L, 1)
	b := luaCheckAccount(L, 2)

	L.Push(lua.LBool(a.AccountId == b.AccountId))

	return 1
}
