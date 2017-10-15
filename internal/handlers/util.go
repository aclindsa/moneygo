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

type ResponseWrapper struct {
	Code   int
	Writer ResponseWriterWriter
}

func (r ResponseWrapper) Write(w http.ResponseWriter) error {
	w.WriteHeader(r.Code)
	return r.Writer.Write(w)
}

type SuccessWriter struct{}

func (s SuccessWriter) Write(w http.ResponseWriter) error {
	fmt.Fprint(w, "{}")
	return nil
}
