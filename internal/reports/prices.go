package reports

import (
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/yuin/gopher-lua"
)

const luaPriceTypeName = "price"

func luaRegisterPrices(L *lua.LState) {
	mt := L.NewTypeMetatable(luaPriceTypeName)
	L.SetGlobal("price", mt)
	L.SetField(mt, "__index", L.NewFunction(luaPrice__index))
	L.SetField(mt, "__tostring", L.NewFunction(luaPrice__tostring))
	L.SetField(mt, "__metatable", lua.LString("protected"))
}

func PriceToLua(L *lua.LState, price *models.Price) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = price
	L.SetMetatable(ud, L.GetTypeMetatable(luaPriceTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Price and returns this *Price.
func luaCheckPrice(L *lua.LState, n int) *models.Price {
	ud := L.CheckUserData(n)
	if price, ok := ud.Value.(*models.Price); ok {
		return price
	}
	L.ArgError(n, "price expected")
	return nil
}

func luaPrice__index(L *lua.LState) int {
	p := luaCheckPrice(L, 1)
	field := L.CheckString(2)

	switch field {
	case "PriceId", "priceid":
		L.Push(lua.LNumber(float64(p.PriceId)))
	case "Security", "security":
		security_map, err := luaContextGetSecurities(L)
		if err != nil {
			panic("luaContextGetSecurities couldn't fetch securities")
		}
		s, ok := security_map[p.SecurityId]
		if !ok {
			panic("Price's security not found for user")
		}
		L.Push(SecurityToLua(L, s))
	case "Currency", "currency":
		security_map, err := luaContextGetSecurities(L)
		if err != nil {
			panic("luaContextGetSecurities couldn't fetch securities")
		}
		c, ok := security_map[p.CurrencyId]
		if !ok {
			panic("Price's currency not found for user")
		}
		L.Push(SecurityToLua(L, c))
	case "Value", "value":
		float, _ := p.Value.Float64()
		L.Push(lua.LNumber(float))
	default:
		L.ArgError(2, "unexpected price attribute: "+field)
	}

	return 1
}

func luaPrice__tostring(L *lua.LState) int {
	p := luaCheckPrice(L, 1)

	security_map, err := luaContextGetSecurities(L)
	if err != nil {
		panic("luaContextGetSecurities couldn't fetch securities")
	}
	s, ok1 := security_map[p.SecurityId]
	c, ok2 := security_map[p.CurrencyId]
	if !ok1 || !ok2 {
		panic("Price's currency or security not found for user")
	}

	L.Push(lua.LString(p.Value.String() + " " + c.Symbol + " (" + s.Symbol + ")"))

	return 1
}
