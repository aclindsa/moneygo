package models

import (
	"encoding/json"
	"net/http"
	"strings"
)

type SecurityType int64

const (
	Currency SecurityType = 1
	Stock                 = 2
)

func GetSecurityType(typestring string) SecurityType {
	if strings.EqualFold(typestring, "currency") {
		return Currency
	} else if strings.EqualFold(typestring, "stock") {
		return Stock
	} else {
		return 0
	}
}

// MaxPrexision denotes the maximum valid value for Security.Precision
const MaxPrecision uint64 = 15

type Security struct {
	SecurityId  int64
	UserId      int64
	Name        string
	Description string
	Symbol      string
	// Number of decimal digits (to the right of the decimal point) this
	// security is precise to
	Precision uint64 `db:"Preciseness"`
	Type      SecurityType
	// AlternateId is CUSIP for Type=Stock, ISO4217 for Type=Currency
	AlternateId string
}

type SecurityList struct {
	Securities *[]*Security `json:"securities"`
}

func (s *Security) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(s)
}

func (s *Security) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(s)
}

func (sl *SecurityList) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(sl)
}

func (sl *SecurityList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(sl)
}
