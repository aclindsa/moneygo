package main

import (
	"github.com/yuin/gopher-lua"
	"math/big"
)

type Balance struct {
	Security *Security
	Amount   *big.Rat
}

const luaBalanceTypeName = "balance"

// Registers my balance type to given L.
func luaRegisterBalances(L *lua.LState) {
	mt := L.NewTypeMetatable(luaBalanceTypeName)
	L.SetGlobal("balance", mt)
	L.SetField(mt, "__tostring", L.NewFunction(luaBalance__tostring))
	L.SetField(mt, "__eq", L.NewFunction(luaBalance__eq))
	L.SetField(mt, "__metatable", lua.LString("protected"))
}

func BalanceToLua(L *lua.LState, balance *Balance) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = balance
	L.SetMetatable(ud, L.GetTypeMetatable(luaBalanceTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Balance and returns this *Balance.
func luaCheckBalance(L *lua.LState, n int) *Balance {
	ud := L.CheckUserData(n)
	if balance, ok := ud.Value.(*Balance); ok {
		return balance
	}
	L.ArgError(n, "balance expected")
	return nil
}

func luaBalance__tostring(L *lua.LState) int {
	b := luaCheckBalance(L, 1)

	L.Push(lua.LString(b.Security.Symbol + b.Amount.FloatString(b.Security.Precision)))

	return 1
}

func luaBalance__eq(L *lua.LState) int {
	a := luaCheckBalance(L, 1)
	b := luaCheckBalance(L, 2)

	L.Push(lua.LBool(a.Security.SecurityId == b.Security.SecurityId && a.Amount.Cmp(b.Amount) == 0))

	return 1
}
