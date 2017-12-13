package reports

import (
	"context"
	"errors"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"github.com/yuin/gopher-lua"
	"strings"
)

const luaAccountTypeName = "account"

func luaContextGetAccounts(L *lua.LState) (map[int64]*models.Account, error) {
	var account_map map[int64]*models.Account

	ctx := L.Context()

	tx, ok := ctx.Value(dbContextKey).(store.Tx)
	if !ok {
		return nil, errors.New("Couldn't find tx in lua's Context")
	}

	account_map, ok = ctx.Value(accountsContextKey).(map[int64]*models.Account)
	if !ok {
		user, ok := ctx.Value(userContextKey).(*models.User)
		if !ok {
			return nil, errors.New("Couldn't find User in lua's Context")
		}

		accounts, err := tx.GetAccounts(user.UserId)
		if err != nil {
			return nil, err
		}

		account_map = make(map[int64]*models.Account)
		for i := range *accounts {
			account_map[(*accounts)[i].AccountId] = (*accounts)[i]
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

func luaRegisterAccounts(L *lua.LState) {
	mt := L.NewTypeMetatable(luaAccountTypeName)
	L.SetGlobal("account", mt)
	L.SetField(mt, "__index", L.NewFunction(luaAccount__index))
	L.SetField(mt, "__tostring", L.NewFunction(luaAccount__tostring))
	L.SetField(mt, "__eq", L.NewFunction(luaAccount__eq))
	L.SetField(mt, "__metatable", lua.LString("protected"))

	for _, accttype := range models.AccountTypes {
		L.SetField(mt, accttype.String(), lua.LNumber(float64(accttype)))
	}

	getAccountsFn := L.NewFunction(luaGetAccounts)
	L.SetField(mt, "get_all", getAccountsFn)
	// also register the get_accounts function as a global in its own right
	L.SetGlobal("get_accounts", getAccountsFn)
}

func AccountToLua(L *lua.LState, account *models.Account) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = account
	L.SetMetatable(ud, L.GetTypeMetatable(luaAccountTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Account and returns this *Account.
func luaCheckAccount(L *lua.LState, n int) *models.Account {
	ud := L.CheckUserData(n)
	if account, ok := ud.Value.(*models.Account); ok {
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
	case "SecurityId", "securityid":
		L.Push(lua.LNumber(float64(a.SecurityId)))
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
	case "TypeName", "Typename":
		L.Push(lua.LString(a.Type.String()))
	case "typename":
		L.Push(lua.LString(strings.ToLower(a.Type.String())))
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
	tx, ok := ctx.Value(dbContextKey).(store.Tx)
	if !ok {
		panic("Couldn't find tx in lua's Context")
	}
	user, ok := ctx.Value(userContextKey).(*models.User)
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
	var balance *models.Amount
	if date != nil {
		end := luaWeakCheckTime(L, 3)
		if end != nil {
			balance, err = tx.GetAccountBalanceDateRange(user, a.AccountId, date, end)
		} else {
			balance, err = tx.GetAccountBalanceDate(user, a.AccountId, date)
		}
	} else {
		balance, err = tx.GetAccountBalance(user, a.AccountId)
	}
	if err != nil {
		panic("Failed to fetch balance for account:" + err.Error())
	}
	b := &Balance{
		Amount:   *balance,
		Security: security,
	}

	L.Push(BalanceToLua(L, b))

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
