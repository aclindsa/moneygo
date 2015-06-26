package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func GetURLID(url string) (int64, error) {
	pieces := strings.Split(strings.Trim(url, "/"), "/")
	return strconv.ParseInt(pieces[len(pieces)-1], 10, 0)
}

func WriteSuccess(w http.ResponseWriter) {
	fmt.Fprint(w, "{}")
}
