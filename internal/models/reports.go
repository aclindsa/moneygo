package models

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Report struct {
	ReportId int64
	UserId   int64
	Name     string
	Lua      string
}

// The maximum length (in bytes) the Lua code may be. This is used to set the
// max size of the database columns (with an added fudge factor)
const LuaMaxLength int = 65536

func (r *Report) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(r)
}

func (r *Report) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(r)
}

type ReportList struct {
	Reports *[]*Report `json:"reports"`
}

func (rl *ReportList) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(rl)
}

func (rl *ReportList) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(rl)
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

func (t *Tabulation) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(t)
}

func (t *Tabulation) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(t)
}
