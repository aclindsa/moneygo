package main

import (
	"github.com/yuin/gopher-lua"
)

const luaReportTypeName = "report"
const luaSeriesTypeName = "series"

func luaRegisterReports(L *lua.LState) {
	mtr := L.NewTypeMetatable(luaReportTypeName)
	L.SetGlobal("report", mtr)
	L.SetField(mtr, "new", L.NewFunction(luaReportNew))
	L.SetField(mtr, "__index", L.NewFunction(luaReport__index))
	L.SetField(mtr, "__metatable", lua.LString("protected"))

	mts := L.NewTypeMetatable(luaSeriesTypeName)
	L.SetGlobal("series", mts)
	L.SetField(mts, "__index", L.NewFunction(luaSeries__index))
	L.SetField(mts, "__metatable", lua.LString("protected"))
}

// Checks whether the first lua argument is a *LUserData with *Report and returns *Report
func luaCheckReport(L *lua.LState, n int) *Report {
	ud := L.CheckUserData(n)
	if report, ok := ud.Value.(*Report); ok {
		return report
	}
	L.ArgError(n, "report expected")
	return nil
}

// Checks whether the first lua argument is a *LUserData with *Series and returns *Series
func luaCheckSeries(L *lua.LState, n int) *Series {
	ud := L.CheckUserData(n)
	if series, ok := ud.Value.(*Series); ok {
		return series
	}
	L.ArgError(n, "series expected")
	return nil
}

func luaReportNew(L *lua.LState) int {
	numvalues := L.CheckInt(1)
	ud := L.NewUserData()
	ud.Value = &Report{
		Labels: make([]string, numvalues),
		Series: make(map[string]*Series),
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaReportTypeName))
	L.Push(ud)
	return 1
}

func luaReport__index(L *lua.LState) int {
	field := L.CheckString(2)

	switch field {
	case "Label", "label":
		L.Push(L.NewFunction(luaReportLabel))
	case "Series", "series":
		L.Push(L.NewFunction(luaReportSeries))
	case "Title", "title":
		L.Push(L.NewFunction(luaReportTitle))
	case "Subtitle", "subtitle":
		L.Push(L.NewFunction(luaReportSubtitle))
	case "XAxisLabel", "xaxislabel":
		L.Push(L.NewFunction(luaReportXAxis))
	case "YAxisLabel", "yaxislabel":
		L.Push(L.NewFunction(luaReportYAxis))
	default:
		L.ArgError(2, "unexpected report attribute: "+field)
	}

	return 1
}

func luaReportLabel(L *lua.LState) int {
	report := luaCheckReport(L, 1)
	labelnumber := L.CheckInt(2)
	label := L.CheckString(3)

	if labelnumber > cap(report.Labels) || labelnumber < 1 {
		L.ArgError(2, "Label index must be between 1 and the number of data points, inclusive")
	}
	report.Labels[labelnumber-1] = label
	return 0
}

func luaReportSeries(L *lua.LState) int {
	report := luaCheckReport(L, 1)
	name := L.CheckString(2)
	ud := L.NewUserData()

	s, ok := report.Series[name]
	if ok {
		ud.Value = s
	} else {
		report.Series[name] = &Series{
			Series: make(map[string]*Series),
			Values: make([]float64, cap(report.Labels)),
		}
		ud.Value = report.Series[name]
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaSeriesTypeName))
	L.Push(ud)
	return 1
}

func luaReportTitle(L *lua.LState) int {
	report := luaCheckReport(L, 1)

	if L.GetTop() == 2 {
		report.Title = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(report.Title))
	return 1
}

func luaReportSubtitle(L *lua.LState) int {
	report := luaCheckReport(L, 1)

	if L.GetTop() == 2 {
		report.Subtitle = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(report.Subtitle))
	return 1
}

func luaReportXAxis(L *lua.LState) int {
	report := luaCheckReport(L, 1)

	if L.GetTop() == 2 {
		report.XAxisLabel = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(report.XAxisLabel))
	return 1
}

func luaReportYAxis(L *lua.LState) int {
	report := luaCheckReport(L, 1)

	if L.GetTop() == 2 {
		report.YAxisLabel = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(report.YAxisLabel))
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
		parent.Series[name] = &Series{
			Series: make(map[string]*Series),
			Values: make([]float64, cap(parent.Values)),
		}
		ud.Value = parent.Series[name]
	}
	L.SetMetatable(ud, L.GetTypeMetatable(luaSeriesTypeName))
	L.Push(ud)
	return 1
}
