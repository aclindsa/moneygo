package main

import (
	"context"
	"github.com/yuin/gopher-lua"
	"log"
	"net/http"
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
				panic(err)
			}
		}

		luaRegisterAccounts(L)
		luaRegisterSecurities(L)
		luaRegisterBalances(L)
		luaRegisterDates(L)

		err := L.DoString(`accounts = account.get_all()
last_parent = nil
for id, account in pairs(accounts) do
	parent = account.parent
	print(account, parent, account.security)
	if parent then
		print(last_parent, parent)
		print("parent equals last:", last_parent == parent)
		last_parent = parent
	end
end
`)

		if err != nil {
			log.Print("lua:" + err.Error())
		}
	}
}
