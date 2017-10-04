package handlers

import (
	"github.com/yuin/gopher-lua"
	"math/big"
)

type Balance struct {
	Security *Security
	Amount   *big.Rat
}

const luaBalanceTypeName = "balance"

func luaRegisterBalances(L *lua.LState) {
	mt := L.NewTypeMetatable(luaBalanceTypeName)
	L.SetGlobal("balance", mt)
	L.SetField(mt, "__index", L.NewFunction(luaBalance__index))
	L.SetField(mt, "__tostring", L.NewFunction(luaBalance__tostring))
	L.SetField(mt, "__eq", L.NewFunction(luaBalance__eq))
	L.SetField(mt, "__lt", L.NewFunction(luaBalance__lt))
	L.SetField(mt, "__le", L.NewFunction(luaBalance__le))
	L.SetField(mt, "__add", L.NewFunction(luaBalance__add))
	L.SetField(mt, "__sub", L.NewFunction(luaBalance__sub))
	L.SetField(mt, "__mul", L.NewFunction(luaBalance__mul))
	L.SetField(mt, "__div", L.NewFunction(luaBalance__div))
	L.SetField(mt, "__unm", L.NewFunction(luaBalance__unm))
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

func luaWeakCheckBalance(L *lua.LState, n int) *Balance {
	v := L.Get(n)
	if ud, ok := v.(*lua.LUserData); ok {
		if balance, ok := ud.Value.(*Balance); ok {
			return balance
		}
		L.ArgError(n, "balance expected")
	}
	return nil
}

func luaGetBalanceOperands(L *lua.LState, n int, m int) (*Balance, *Balance) {
	bn := luaWeakCheckBalance(L, n)
	bm := luaWeakCheckBalance(L, m)

	if bn != nil && bm != nil {
		return bn, bm
	} else if bn != nil {
		nm := L.CheckNumber(m)
		var balance Balance
		var rat big.Rat
		balance.Security = bn.Security
		balance.Amount = rat.SetFloat64(float64(nm))
		if balance.Amount == nil {
			L.ArgError(n, "non-finite float invalid for operand to balance arithemetic")
			return nil, nil
		}
		return bn, &balance
	} else if bm != nil {
		nn := L.CheckNumber(n)
		var balance Balance
		var rat big.Rat
		balance.Security = bm.Security
		balance.Amount = rat.SetFloat64(float64(nn))
		if balance.Amount == nil {
			L.ArgError(n, "non-finite float invalid for operand to balance arithemetic")
			return nil, nil
		}
		return bm, &balance
	}
	L.ArgError(n, "balance expected")
	return nil, nil
}

func luaBalance__index(L *lua.LState) int {
	a := luaCheckBalance(L, 1)
	field := L.CheckString(2)

	switch field {
	case "Security", "security":
		L.Push(SecurityToLua(L, a.Security))
	case "Amount", "amount":
		float, _ := a.Amount.Float64()
		L.Push(lua.LNumber(float))
	default:
		L.ArgError(2, "unexpected balance attribute: "+field)
	}

	return 1
}

func luaBalance__tostring(L *lua.LState) int {
	b := luaCheckBalance(L, 1)

	L.Push(lua.LString(b.Security.Symbol + " " + b.Amount.FloatString(b.Security.Precision)))

	return 1
}

func luaBalance__eq(L *lua.LState) int {
	a := luaCheckBalance(L, 1)
	b := luaCheckBalance(L, 2)

	L.Push(lua.LBool(a.Security.SecurityId == b.Security.SecurityId && a.Amount.Cmp(b.Amount) == 0))

	return 1
}

func luaBalance__lt(L *lua.LState) int {
	a := luaCheckBalance(L, 1)
	b := luaCheckBalance(L, 2)
	if a.Security.SecurityId != b.Security.SecurityId {
		L.ArgError(2, "Can't compare balances with different securities")
	}

	L.Push(lua.LBool(a.Amount.Cmp(b.Amount) < 0))

	return 1
}

func luaBalance__le(L *lua.LState) int {
	a := luaCheckBalance(L, 1)
	b := luaCheckBalance(L, 2)
	if a.Security.SecurityId != b.Security.SecurityId {
		L.ArgError(2, "Can't compare balances with different securities")
	}

	L.Push(lua.LBool(a.Amount.Cmp(b.Amount) <= 0))

	return 1
}

func luaBalance__add(L *lua.LState) int {
	a, b := luaGetBalanceOperands(L, 1, 2)

	if a.Security.SecurityId != b.Security.SecurityId {
		L.ArgError(2, "Can't add balances with different securities")
	}

	var balance Balance
	var rat big.Rat
	balance.Security = a.Security
	balance.Amount = rat.Add(a.Amount, b.Amount)
	L.Push(BalanceToLua(L, &balance))

	return 1
}

func luaBalance__sub(L *lua.LState) int {
	a, b := luaGetBalanceOperands(L, 1, 2)

	if a.Security.SecurityId != b.Security.SecurityId {
		L.ArgError(2, "Can't subtract balances with different securities")
	}

	var balance Balance
	var rat big.Rat
	balance.Security = a.Security
	balance.Amount = rat.Sub(a.Amount, b.Amount)
	L.Push(BalanceToLua(L, &balance))

	return 1
}

func luaBalance__mul(L *lua.LState) int {
	a, b := luaGetBalanceOperands(L, 1, 2)

	if a.Security.SecurityId != b.Security.SecurityId {
		L.ArgError(2, "Can't multiply balances with different securities")
	}

	var balance Balance
	var rat big.Rat
	balance.Security = a.Security
	balance.Amount = rat.Mul(a.Amount, b.Amount)
	L.Push(BalanceToLua(L, &balance))

	return 1
}

func luaBalance__div(L *lua.LState) int {
	a, b := luaGetBalanceOperands(L, 1, 2)

	if a.Security.SecurityId != b.Security.SecurityId {
		L.ArgError(2, "Can't divide balances with different securities")
	}

	var balance Balance
	var rat big.Rat
	balance.Security = a.Security
	balance.Amount = rat.Quo(a.Amount, b.Amount)
	L.Push(BalanceToLua(L, &balance))

	return 1
}

func luaBalance__unm(L *lua.LState) int {
	b := luaCheckBalance(L, 1)

	var balance Balance
	var rat big.Rat
	balance.Security = b.Security
	balance.Amount = rat.Neg(b.Amount)
	L.Push(BalanceToLua(L, &balance))

	return 1
}
