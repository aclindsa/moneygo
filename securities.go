package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	Banknote   int64 = 1
	Bond             = 2
	Stock            = 3
	MutualFund       = 4
)

type Security struct {
	SecurityId int64
	Name       string
	// Number of decimal digits (to the right of the decimal point) this
	// security is precise to
	Precision int64
	Type      int64
}

type SecurityList struct {
	Securities *[]*Security `json:"securities"`
}

var security_map = map[int64]*Security{
	1: &Security{
		SecurityId: 1,
		Name:       "USD",
		Precision:  2,
		Type:       Banknote},
	2: &Security{
		SecurityId: 2,
		Name:       "SPY",
		Precision:  5,
		Type:       Stock},
}

var security_list []*Security

func init() {
	for _, value := range security_map {
		security_list = append(security_list, value)
	}
}

func GetSecurity(securityid int64) *Security {
	s := security_map[securityid]
	if s != nil {
		return s
	}
	return nil
}

func GetSecurities() []*Security {
	return security_list
}

func (s *Security) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func (sl *SecurityList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(sl)
}

func SecurityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		securityid, err := GetURLID(r.URL.Path)
		if err == nil {
			security := GetSecurity(securityid)
			if security == nil {
				WriteError(w, 3 /*Invalid Request*/)
				return
			}
			err := security.Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		} else {
			var sl SecurityList
			securities := GetSecurities()
			sl.Securities = &securities
			err := (&sl).Write(w)
			if err != nil {
				WriteError(w, 999 /*Internal Error*/)
				log.Print(err)
				return
			}
		}
	} else {
		WriteError(w, 3 /*Invalid Request*/)
	}
}
