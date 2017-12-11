package reports

import (
	"context"
	"errors"
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/store"
	"github.com/yuin/gopher-lua"
	"time"
)

//type and value to store user in lua's Context
type key int

const (
	userContextKey key = iota
	accountsContextKey
	securitiesContextKey
	balanceContextKey
	dbContextKey
)

const luaTimeoutSeconds time.Duration = 30 // maximum time a lua request can run for

func RunReport(tx store.Tx, user *models.User, report *models.Report) (*models.Tabulation, error) {
	// Create a new LState without opening the default libs for security
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	// Create a new context holding the current user with a timeout
	ctx := context.WithValue(context.Background(), userContextKey, user)
	ctx = context.WithValue(ctx, dbContextKey, tx)
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
	luaRegisterTabulations(L)
	luaRegisterPrices(L)

	err := L.DoString(report.Lua)

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
		if tabulation, ok := ud.Value.(*models.Tabulation); ok {
			return tabulation, nil
		} else {
			return nil, fmt.Errorf("generate() for %s (Id: %d) didn't return a tabulation", report.Name, report.ReportId)
		}
	} else {
		return nil, fmt.Errorf("generate() for %s (Id: %d) didn't even return LUserData", report.Name, report.ReportId)
	}
}
