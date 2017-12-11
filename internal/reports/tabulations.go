package reports

import (
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/yuin/gopher-lua"
)

const luaTabulationTypeName = "tabulation"
const luaSeriesTypeName = "series"

func luaRegisterTabulations(L *lua.LState) {
	mtr := L.NewTypeMetatable(luaTabulationTypeName)
	L.SetGlobal("tabulation", mtr)
	L.SetField(mtr, "new", L.NewFunction(luaTabulationNew))
	L.SetField(mtr, "__index", L.NewFunction(luaTabulation__index))
	L.SetField(mtr, "__metatable", lua.LString("protected"))

	mts := L.NewTypeMetatable(luaSeriesTypeName)
	L.SetGlobal("series", mts)
	L.SetField(mts, "__index", L.NewFunction(luaSeries__index))
	L.SetField(mts, "__metatable", lua.LString("protected"))
}

// Checks whether the first lua argument is a *LUserData with *Tabulation and returns *Tabulation
func luaCheckTabulation(L *lua.LState, n int) *models.Tabulation {
	ud := L.CheckUserData(n)
	if tabulation, ok := ud.Value.(*models.Tabulation); ok {
		return tabulation
	}
	L.ArgError(n, "tabulation expected")
	return nil
}

// Checks whether the first lua argument is a *LUserData with *Series and returns *Series
func luaCheckSeries(L *lua.LState, n int) *models.Series {
	ud := L.CheckUserData(n)
	if series, ok := ud.Value.(*models.Series); ok {
		return series
	}
	L.ArgError(n, "series expected")
	return nil
}

func luaTabulationNew(L *lua.LState) int {
	numvalues := L.CheckInt(1)
	ud := L.NewUserData()
	ud.Value = &models.Tabulation{
		Labels: make([]string, numvalues),
		Series: make(map[string]*models.Series),
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaTabulationTypeName))
	L.Push(ud)
	return 1
}

func luaTabulation__index(L *lua.LState) int {
	field := L.CheckString(2)

	switch field {
	case "Label", "label":
		L.Push(L.NewFunction(luaTabulationLabel))
	case "Series", "series":
		L.Push(L.NewFunction(luaTabulationSeries))
	case "Title", "title":
		L.Push(L.NewFunction(luaTabulationTitle))
	case "Subtitle", "subtitle":
		L.Push(L.NewFunction(luaTabulationSubtitle))
	case "Units", "units":
		L.Push(L.NewFunction(luaTabulationUnits))
	default:
		L.ArgError(2, "unexpected tabulation attribute: "+field)
	}

	return 1
}

func luaTabulationLabel(L *lua.LState) int {
	tabulation := luaCheckTabulation(L, 1)
	labelnumber := L.CheckInt(2)
	label := L.CheckString(3)

	if labelnumber > cap(tabulation.Labels) || labelnumber < 1 {
		L.ArgError(2, "Label index must be between 1 and the number of data points, inclusive")
	}
	tabulation.Labels[labelnumber-1] = label
	return 0
}

func luaTabulationSeries(L *lua.LState) int {
	tabulation := luaCheckTabulation(L, 1)
	name := L.CheckString(2)
	ud := L.NewUserData()

	s, ok := tabulation.Series[name]
	if ok {
		ud.Value = s
	} else {
		tabulation.Series[name] = &models.Series{
			Series: make(map[string]*models.Series),
			Values: make([]float64, cap(tabulation.Labels)),
		}
		ud.Value = tabulation.Series[name]
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaSeriesTypeName))
	L.Push(ud)
	return 1
}

func luaTabulationTitle(L *lua.LState) int {
	tabulation := luaCheckTabulation(L, 1)

	if L.GetTop() == 2 {
		tabulation.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(tabulation.Title))
	return 1
}

func luaTabulationSubtitle(L *lua.LState) int {
	tabulation := luaCheckTabulation(L, 1)

	if L.GetTop() == 2 {
		tabulation.Subtitle = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(tabulation.Subtitle))
	return 1
}

func luaTabulationUnits(L *lua.LState) int {
	tabulation := luaCheckTabulation(L, 1)

	if L.GetTop() == 2 {
		tabulation.Units = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(tabulation.Units))
	return 1
}

func luaSeries__index(L *lua.LState) int {
	field := L.CheckString(2)

	switch field {
	case "Value", "value":
		L.Push(L.NewFunction(luaSeriesValue))
	case "Series", "series":
		L.Push(L.NewFunction(luaSeriesSeries))
	default:
		L.ArgError(2, "unexpected series attribute: "+field)
	}

	return 1
}

func luaSeriesValue(L *lua.LState) int {
	series := luaCheckSeries(L, 1)
	valuenumber := L.CheckInt(2)
	value := float64(L.CheckNumber(3))

	if valuenumber > cap(series.Values) || valuenumber < 1 {
		L.ArgError(2, "value index must be between 1 and the number of data points, inclusive")
	}
	series.Values[valuenumber-1] = value

	return 0
}

func luaSeriesSeries(L *lua.LState) int {
	parent := luaCheckSeries(L, 1)
	name := L.CheckString(2)
	ud := L.NewUserData()

	s, ok := parent.Series[name]
	if ok {
		ud.Value = s
	} else {
		parent.Series[name] = &models.Series{
			Series: make(map[string]*models.Series),
			Values: make([]float64, cap(parent.Values)),
		}
		ud.Value = parent.Series[name]
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaSeriesTypeName))
	L.Push(ud)
	return 1
}
