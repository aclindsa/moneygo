package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Error struct {
	ErrorId     int
	ErrorString string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error %d: %s", e.ErrorId, e.ErrorString)
}

func (e *Error) Read(json_str string) error {
	dec := json.NewDecoder(strings.NewReader(json_str))
	return dec.Decode(e)
}

func (e *Error) Write(w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(e)
}

var error_codes = map[int]string{
	1: "Not Signed In",
	2: "Unauthorized Access",
	3: "Invalid Request",
	4: "User Exists",
	//  5:   "Connection Failed", //reserved for client-side error
	6:   "Import Error",
	7:   "In Use Error",
	999: "Internal Error",
}

func NewError(error_code int) *Error {
	msg, ok := error_codes[error_code]
	if !ok {
		log.Printf("Error: NewError received unknown error code of %d", error_code)
		msg = error_codes[999]
	}
	return &Error{error_code, msg}
}
