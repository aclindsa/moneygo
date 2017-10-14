package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yuin/gopher-lua"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var reportTabulationRE *regexp.Regexp

func init() {
	reportTabulationRE = regexp.MustCompile(`^/report/[0-9]+/tabulation/?$`)
}

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

type Report struct {
	ReportId int64
	UserId   int64
	Name     string
	Lua      string
}

func (r *Report) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(r)
}

func (r *Report) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(r)
}

type ReportList struct {
	Reports *[]Report `json:"reports"`
}

func (rl *ReportList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(rl)
}

type Series struct {
	Values []float64
	Series map[string]*Series
}

type Tabulation struct {
	ReportId int64
	Title    string
	Subtitle string
	Units    string
	Labels   []string
	Series   map[string]*Series
}

func (r *Tabulation) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(r)
}

func GetReport(tx *Tx, reportid int64, userid int64) (*Report, error) {
	var r Report

	err := tx.SelectOne(&r, "SELECT * from reports where UserId=? AND ReportId=?", userid, reportid)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func GetReports(tx *Tx, userid int64) (*[]Report, error) {
	var reports []Report

	_, err := tx.Select(&reports, "SELECT * from reports where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &reports, nil
}

func InsertReport(tx *Tx, r *Report) error {
	err := tx.Insert(r)
	if err != nil {
		return err
	}
	return nil
}

func UpdateReport(tx *Tx, r *Report) error {
	count, err := tx.Update(r)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Updated more than one report")
	}
	return nil
}

func DeleteReport(tx *Tx, r *Report) error {
	count, err := tx.Delete(r)
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("Deleted more than one report")
	}
	return nil
}

func runReport(tx *Tx, user *User, report *Report) (*Tabulation, error) {
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
		if tabulation, ok := ud.Value.(*Tabulation); ok {
			return tabulation, nil
		} else {
			return nil, fmt.Errorf("generate() for %s (Id: %d) didn't return a tabulation", report.Name, report.ReportId)
		}
	} else {
		return nil, fmt.Errorf("generate() for %s (Id: %d) didn't even return LUserData", report.Name, report.ReportId)
	}
}

func ReportTabulationHandler(tx *Tx, r *http.Request, user *User, reportid int64) ResponseWriterWriter {
	report, err := GetReport(tx, reportid, user.UserId)
	if err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	tabulation, err := runReport(tx, user, report)
	if err != nil {
		// TODO handle different failure cases differently
		log.Print("runReport returned:", err)
		return NewError(3 /*Invalid Request*/)
	}

	tabulation.ReportId = reportid

	return tabulation
}

func ReportHandler(r *http.Request, tx *Tx) ResponseWriterWriter {
	user, err := GetUserFromSession(tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method == "POST" {
		report_json := r.PostFormValue("report")
		if report_json == "" {
			return NewError(3 /*Invalid Request*/)
		}

		var report Report
		err := report.Read(report_json)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		report.ReportId = -1
		report.UserId = user.UserId

		err = InsertReport(tx, &report)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		return ResponseWrapper{201, &report}
	} else if r.Method == "GET" {
		if reportTabulationRE.MatchString(r.URL.Path) {
			var reportid int64
			n, err := GetURLPieces(r.URL.Path, "/report/%d/tabulation", &reportid)
			if err != nil || n != 1 {
				log.Print(err)
				return NewError(999 /*InternalError*/)
			}
			return ReportTabulationHandler(tx, r, user, reportid)
		}

		var reportid int64
		n, err := GetURLPieces(r.URL.Path, "/report/%d", &reportid)
		if err != nil || n != 1 {
			//Return all Reports
			var rl ReportList
			reports, err := GetReports(tx, user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			rl.Reports = reports
			return &rl
		} else {
			// Return Report with this Id
			report, err := GetReport(tx, reportid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			return report
		}
	} else {
		reportid, err := GetURLID(r.URL.Path)
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		if r.Method == "PUT" {
			report_json := r.PostFormValue("report")
			if report_json == "" {
				return NewError(3 /*Invalid Request*/)
			}

			var report Report
			err := report.Read(report_json)
			if err != nil || report.ReportId != reportid {
				return NewError(3 /*Invalid Request*/)
			}
			report.UserId = user.UserId

			err = UpdateReport(tx, &report)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &report
		} else if r.Method == "DELETE" {
			report, err := GetReport(tx, reportid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = DeleteReport(tx, report)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
