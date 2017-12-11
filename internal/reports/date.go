package reports

import (
	"github.com/yuin/gopher-lua"
	"time"
)

const luaDateTypeName = "date"
const timeFormat = "2006-01-02"

func luaRegisterDates(L *lua.LState) {
	mt := L.NewTypeMetatable(luaDateTypeName)
	L.SetGlobal("date", mt)
	L.SetField(mt, "new", L.NewFunction(luaDateNew))
	L.SetField(mt, "now", L.NewFunction(luaDateNow))
	L.SetField(mt, "__index", L.NewFunction(luaDate__index))
	L.SetField(mt, "__tostring", L.NewFunction(luaDate__tostring))
	L.SetField(mt, "__eq", L.NewFunction(luaDate__eq))
	L.SetField(mt, "__lt", L.NewFunction(luaDate__lt))
	L.SetField(mt, "__le", L.NewFunction(luaDate__le))
	L.SetField(mt, "__add", L.NewFunction(luaDate__add))
	L.SetField(mt, "__sub", L.NewFunction(luaDate__sub))
	L.SetField(mt, "__metatable", lua.LString("protected"))
}

func TimeToLua(L *lua.LState, date *time.Time) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = date
	L.SetMetatable(ud, L.GetTypeMetatable(luaDateTypeName))
	return ud
}

// Checks whether the first lua argument is a *LUserData with *Time and returns this *Time.
func luaCheckTime(L *lua.LState, n int) *time.Time {
	ud := L.CheckUserData(n)
	if date, ok := ud.Value.(*time.Time); ok {
		return date
	}
	L.ArgError(n, "date expected")
	return nil
}

func luaWeakCheckTime(L *lua.LState, n int) *time.Time {
	v := L.Get(n)
	if ud, ok := v.(*lua.LUserData); ok {
		if date, ok := ud.Value.(*time.Time); ok {
			return date
		}
	}
	return nil
}

func luaWeakCheckTableFieldInt(L *lua.LState, T *lua.LTable, n int, name string, def int) int {
	lv := T.RawGetString(name)
	if lv == lua.LNil {
		return def
	}
	if i, ok := lv.(lua.LNumber); ok {
		return int(i)
	}
	L.ArgError(n, "table field '"+name+"' expected to be int")
	return def
}

func luaDateNew(L *lua.LState) int {
	// TODO make this track the user's timezone
	v := L.Get(1)
	if s, ok := v.(lua.LString); ok {
		date, err := time.ParseInLocation(timeFormat, s.String(), time.Local)
		if err != nil {
			L.ArgError(1, "error parsing date string: "+err.Error())
			return 0
		}
		L.Push(TimeToLua(L, &date))
		return 1
	}
	var year, month, day int
	if t, ok := v.(*lua.LTable); ok {
		year = luaWeakCheckTableFieldInt(L, t, 1, "year", 0)
		month = luaWeakCheckTableFieldInt(L, t, 1, "month", 1)
		day = luaWeakCheckTableFieldInt(L, t, 1, "day", 1)
	} else {
		year = L.CheckInt(1)
		month = L.CheckInt(2)
		day = L.CheckInt(3)
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	L.Push(TimeToLua(L, &date))
	return 1
}

func luaDateNow(L *lua.LState) int {
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	L.Push(TimeToLua(L, &date))
	return 1
}

func luaDate__index(L *lua.LState) int {
	d := luaCheckTime(L, 1)
	field := L.CheckString(2)

	switch field {
	case "Year", "year":
		L.Push(lua.LNumber(d.Year()))
	case "Month", "month":
		L.Push(lua.LNumber(float64(d.Month())))
	case "Day", "day":
		L.Push(lua.LNumber(float64(d.Day())))
	default:
		L.ArgError(2, "unexpected date attribute: "+field)
	}

	return 1
}

func luaDate__tostring(L *lua.LState) int {
	a := luaCheckTime(L, 1)

	L.Push(lua.LString(a.Format(timeFormat)))

	return 1
}

func luaDate__eq(L *lua.LState) int {
	a := luaCheckTime(L, 1)
	b := luaCheckTime(L, 2)

	L.Push(lua.LBool(a.Equal(*b)))

	return 1
}

func luaDate__lt(L *lua.LState) int {
	a := luaCheckTime(L, 1)
	b := luaCheckTime(L, 2)

	L.Push(lua.LBool(a.Before(*b)))

	return 1
}

func luaDate__le(L *lua.LState) int {
	a := luaCheckTime(L, 1)
	b := luaCheckTime(L, 2)

	L.Push(lua.LBool(a.Equal(*b) || a.Before(*b)))

	return 1
}

func luaDate__add(L *lua.LState) int {
	a := luaCheckTime(L, 1)
	b := luaCheckTime(L, 2)

	date := a.AddDate(b.Year(), int(b.Month()), b.Day())
	L.Push(TimeToLua(L, &date))

	return 1
}

func luaDate__sub(L *lua.LState) int {
	a := luaCheckTime(L, 1)
	b := luaCheckTime(L, 2)

	date := a.AddDate(-b.Year(), -int(b.Month()), -b.Day())
	L.Push(TimeToLua(L, &date))

	return 1
}
