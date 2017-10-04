package handlers

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

func GetURLPieces(url string, format string, a ...interface{}) (int, error) {
	url = strings.Replace(url, "/", " ", -1)
	format = strings.Replace(format, "/", " ", -1)
	return fmt.Sscanf(url, format, a...)
}

func WriteSuccess(w http.ResponseWriter) {
	fmt.Fprint(w, "{}")
}
