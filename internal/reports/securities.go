package reports

import (
	"context"
	"errors"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"github.com/yuin/gopher-lua"
	"time"
)

const luaSecurityTypeName = "security"

func luaContextGetSecurities(L *lua.LState) (map[int64]*models.Security, error) {
	var security_map map[int64]*models.Security

	ctx := L.Context()

	tx, ok := ctx.Value(dbContextKey).(store.Tx)
	if !ok {
		return nil, errors.New("Couldn't find tx in lua's Context")
	}

	security_map, ok = ctx.Value(securitiesContextKey).(map[int64]*models.Security)
	if !ok {
		user, ok := ctx.Value(userContextKey).(*models.User)
		if !ok {
			return nil, errors.New("Couldn't find User in lua's Context")
		}

		securities, err := tx.GetSecurities(user.UserId)
		if err != nil {
			return nil, err
		}

		security_map = make(map[int64]*models.Security)
		for i := range *securities {
			security_map[(*securities)[i].SecurityId] = (*securities)[i]
		}

		ctx = context.WithValue(ctx, securitiesContextKey, security_map)
		L.SetContext(ctx)
	}

	return security_map, nil
}

func luaContextGetDefaultCurrency(L *lua.LState) (*models.Security, error) {
	security_map, err := luaContextGetSecurities(L)
	if err != nil {
		return nil, err
	}

	ctx := L.Context()

	user, ok := ctx.Value(userContextKey).(*models.User)
	if !ok {
		return nil, errors.New("Couldn't find User in lua's Context")
	}

	if security, ok := security_map[user.DefaultCurrency]; ok {
		return security, nil
	} else {
		return nil, errors.New("DefaultCurrency not in lua security_map")
	}
}

func luaGetDefaultCurrency(L *lua.LState) int {
	defcurrency, err := luaContextGetDefaultCurrency(L)
	if err != nil {
		panic("luaGetDefaultCurrency couldn't fetch default currency")
	}

	L.Push(SecurityToLua(L, defcurrency))
	return 1
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
	getDefaultCurrencyFn := L.NewFunction(luaGetDefaultCurrency)
	L.SetField(mt, "get_default", getDefaultCurrencyFn)

	// also register the get_securities and get_default functions as globals in
	// their own right
	L.SetGlobal("get_securities", getSecuritiesFn)
	L.SetGlobal("get_default_currency", getDefaultCurrencyFn)
}

func SecurityToLua(L *lua.LState, security *models.Security) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = security
	L.SetMetatable(ud, L.GetTypeMetatable(luaSecurityTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Security and returns this *Security.
func luaCheckSecurity(L *lua.LState, n int) *models.Security {
	ud := L.CheckUserData(n)
	if security, ok := ud.Value.(*models.Security); ok {
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
	case "ClosestPrice", "closestprice":
		L.Push(L.NewFunction(luaClosestPrice))
	case "AlternateId", "alternateid":
		L.Push(lua.LString(a.AlternateId))
	default:
		L.ArgError(2, "unexpected security attribute: "+field)
	}

	return 1
}

// Return the price for security in currency closest to date
func getClosestPrice(tx store.Tx, security, currency *models.Security, date *time.Time) (*models.Price, error) {
	earliest, _ := tx.GetEarliestPrice(security, currency, date)
	latest, err := tx.GetLatestPrice(security, currency, date)

	// Return early if either earliest or latest are invalid
	if earliest == nil {
		return latest, err
	} else if err != nil {
		return earliest, nil
	}

	howlate := earliest.Date.Sub(*date)
	howearly := date.Sub(latest.Date)
	if howearly < howlate {
		return latest, nil
	} else {
		return earliest, nil
	}
}

func luaClosestPrice(L *lua.LState) int {
	s := luaCheckSecurity(L, 1)
	c := luaCheckSecurity(L, 2)
	date := luaCheckTime(L, 3)

	ctx := L.Context()
	tx, ok := ctx.Value(dbContextKey).(store.Tx)
	if !ok {
		panic("Couldn't find tx in lua's Context")
	}

	p, err := getClosestPrice(tx, s, c, date)
	if err != nil {
		L.Push(lua.LNil)
	} else {
		L.Push(PriceToLua(L, p))
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
