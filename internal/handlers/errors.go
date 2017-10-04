package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type Error struct {
	ErrorId     int
	ErrorString string
}

var error_codes = map[int]string{
	1: "Not Signed In",
	2: "Unauthorized Access",
	3: "Invalid Request",
	4: "User Exists",
	//  5:   "Connection Failed", //reserved for client-side error
	6:   "Import Error",
	999: "Internal Error",
}

func WriteError(w http.ResponseWriter, error_code int) {
	msg, ok := error_codes[error_code]
	if !ok {
		log.Printf("Error: WriteError received error code of %d", error_code)
		msg = error_codes[999]
	}
	e := Error{error_code, msg}

	enc := json.NewEncoder(w)
	err := enc.Encode(e)
	if err != nil {
		log.Fatal(err)
	}
}
