package models

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type Price struct {
	PriceId    int64
	SecurityId int64
	CurrencyId int64
	Date       time.Time
	Value      Amount // price of Security in Currency units
	RemoteId   string // unique ID from source, for detecting duplicates
}

type PriceList struct {
	Prices *[]*Price `json:"prices"`
}

func (p *Price) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(p)
}

func (p *Price) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(p)
}

func (pl *PriceList) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(pl)
}

func (pl *PriceList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(pl)
}
