package handlers

import (
	"fmt"
	"net/http"
)

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
