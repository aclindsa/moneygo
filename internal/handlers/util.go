package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func ReadJSON(r *http.Request, v interface{}) error {
	jsonstring, err := ioutil.ReadAll(io.LimitReader(r.Body, 10*1024*1024 /*10Mb*/))
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonstring, v)
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
