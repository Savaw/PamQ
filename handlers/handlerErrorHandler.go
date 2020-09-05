package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RootHandler func(http.ResponseWriter, *http.Request) error

func (fn RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err == nil {
		return
	}

	log.Println(err)
}

type ErrorMissingField string

func (e ErrorMissingField) Error() string {
	return string(e) + " is required."
}

type ClientError interface {
	Error() string
	ResponseBody() ([]byte, error)
	ResponseHeaders() (int, map[string]string)
}

type HTTPError struct {
	Cause  error
	Detail string
	Status int
}

func (e *HTTPError) Error() string {
	return e.Detail + ": " + e.Cause.Error()
}

func (e *HTTPError) ResponseBody() ([]byte, error) {
	js, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing response body: %v", err)
	}
	return js, nil
}

func (e *HTTPError) ResponseHeaders() (int, map[string]string) {
	mp := map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
	return e.Status, mp
}

func NewHTTPError(err error, status int, detail string) error {
	return &HTTPError{
		Cause:  err,
		Detail: detail,
		Status: status,
	}
}
