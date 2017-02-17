package main

import (
	"context"
	"errors"
	"github.com/yuin/gopher-lua"
)

const luaSecurityTypeName = "security"

func luaContextGetSecurities(L *lua.LState) (map[int64]*Security, error) {
	var security_map map[int64]*Security

	ctx := L.Context()

	security_map, ok := ctx.Value(securitiesContextKey).(map[int64]*Security)
	if !ok {
		user, ok := ctx.Value(userContextKey).(*User)
		if !ok {
			return nil, errors.New("Couldn't find User in lua's Context")
		}

		securities, err := GetSecurities(user.UserId)
		if err != nil {
			return nil, err
		}

		security_map = make(map[int64]*Security)
		for i := range *securities {
			security_map[(*securities)[i].SecurityId] = (*securities)[i]
		}

		ctx = context.WithValue(ctx, securitiesContextKey, security_map)
		L.SetContext(ctx)
	}

	return security_map, nil
}

func luaGetSecurities(L *lua.LState) int {
	security_map, err := luaContextGetSecurities(L)
	if err != nil {
		panic("luaGetSecurities couldn't fetch securities")
	}

	table := L.NewTable()

	for securityid := range security_map {
		table.RawSetInt(int(securityid), SecurityToLua(L, security_map[securityid]))
	}

	L.Push(table)
	return 1
}

func luaRegisterSecurities(L *lua.LState) {
	mt := L.NewTypeMetatable(luaSecurityTypeName)
	L.SetGlobal("security", mt)
	L.SetField(mt, "__index", L.NewFunction(luaSecurity__index))
	L.SetField(mt, "__tostring", L.NewFunction(luaSecurity__tostring))
	L.SetField(mt, "__eq", L.NewFunction(luaSecurity__eq))
	L.SetField(mt, "__metatable", lua.LString("protected"))
	getSecuritiesFn := L.NewFunction(luaGetSecurities)
	L.SetField(mt, "get_all", getSecuritiesFn)

	// also register the get_securities function as a global in its own right
	L.SetGlobal("get_securities", getSecuritiesFn)
}

func SecurityToLua(L *lua.LState, security *Security) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = security
	L.SetMetatable(ud, L.GetTypeMetatable(luaSecurityTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Security and returns this *Security.
func luaCheckSecurity(L *lua.LState, n int) *Security {
	ud := L.CheckUserData(n)
	if security, ok := ud.Value.(*Security); ok {
		return security
	}
	L.ArgError(n, "security expected")
	return nil
}

func luaSecurity__index(L *lua.LState) int {
	a := luaCheckSecurity(L, 1)
	field := L.CheckString(2)

	switch field {
	case "SecurityId", "securityid":
		L.Push(lua.LNumber(float64(a.SecurityId)))
	case "Name", "name":
		L.Push(lua.LString(a.Name))
	case "Description", "description":
		L.Push(lua.LString(a.Description))
	case "Symbol", "symbol":
		L.Push(lua.LString(a.Symbol))
	case "Precision", "precision":
		L.Push(lua.LNumber(float64(a.Precision)))
	case "Type", "type":
		L.Push(lua.LNumber(float64(a.Type)))
	default:
		L.ArgError(2, "unexpected security attribute: "+field)
	}

	return 1
}

func luaSecurity__tostring(L *lua.LState) int {
	s := luaCheckSecurity(L, 1)

	L.Push(lua.LString(s.Name + " - " + s.Description + " (" + s.Symbol + ")"))

	return 1
}

func luaSecurity__eq(L *lua.LState) int {
	a := luaCheckSecurity(L, 1)
	b := luaCheckSecurity(L, 2)

	L.Push(lua.LBool(a.SecurityId == b.SecurityId))

	return 1
}
