package main

import (
	"context"
	"github.com/yuin/gopher-lua"
	"log"
	"net/http"
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

const luaTimeoutSeconds time.Duration = 5 // maximum time a lua request can run for

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
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		}

		luaRegisterAccounts(L)
		luaRegisterSecurities(L)
		luaRegisterBalances(L)
		luaRegisterDates(L)

		err = L.DoFile(reportpath)

		if err != nil {
			WriteError(w, 3 /*Invalid Request*/)
			log.Print("lua:" + err.Error())
			return
		}
	}
}
