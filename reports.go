package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yuin/gopher-lua"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

//type and value to store user in lua's Context
type key int

const (
	userContextKey key = iota
	accountsContextKey
	securitiesContextKey
	balanceContextKey
)

const luaTimeoutSeconds time.Duration = 30 // maximum time a lua request can run for

type Series struct {
	Values []float64
	Series map[string]*Series
}

type Report struct {
	ReportId   string
	Title      string
	Subtitle   string
	XAxisLabel string
	YAxisLabel string
	Labels     []string
	Series     map[string]*Series
}

func (r *Report) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(r)
}

func runReport(user *User, reportpath string) (*Report, error) {
	// Create a new LState without opening the default libs for security
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	// Create a new context holding the current user with a timeout
	ctx := context.WithValue(context.Background(), userContextKey, user)
	ctx, cancel := context.WithTimeout(ctx, luaTimeoutSeconds*time.Second)
	defer cancel()
	L.SetContext(ctx)

	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			return nil, errors.New("Error initializing Lua packages")
		}
	}

	luaRegisterAccounts(L)
	luaRegisterSecurities(L)
	luaRegisterBalances(L)
	luaRegisterDates(L)
	luaRegisterReports(L)

	err := L.DoFile(reportpath)

	if err != nil {
		return nil, err
	}

	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("generate"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return nil, err
	}

	value := L.Get(-1)
	if ud, ok := value.(*lua.LUserData); ok {
		if report, ok := ud.Value.(*Report); ok {
			return report, nil
		} else {
			return nil, errors.New("generate() in " + reportpath + " didn't return a report")
		}
	} else {
		return nil, errors.New("generate() in " + reportpath + " didn't return a report")
	}
}

func ReportHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		WriteError(w, 1 /*Not Signed In*/)
		return
	}

	if r.Method == "GET" {
		var reportname string
		n, err := GetURLPieces(r.URL.Path, "/report/%s", &reportname)
		if err != nil || n != 1 {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		reportpath := path.Join(baseDir, "reports", reportname+".lua")
		report_stat, err := os.Stat(reportpath)
		if err != nil || !report_stat.Mode().IsRegular() {
			WriteError(w, 3 /*Invalid Request*/)
			return
		}

		report, err := runReport(user, reportpath)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
		report.ReportId = reportname

		err = report.Write(w)
		if err != nil {
			WriteError(w, 999 /*Internal Error*/)
			log.Print(err)
			return
		}
	}
}
