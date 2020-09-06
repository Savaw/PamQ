package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type RootHandler func(http.ResponseWriter, *http.Request) error

func (fn RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err == nil {
		return
	}
	log.Printf("An error accured: %v %T", err, err)

	httpError, ok := err.(*HTTPError)
	if !ok {
		log.Println("Not http error")
		w.WriteHeader(500)
		return
	}

	switch httpError.Type {
	case ClientError:
		pqErr, ok := httpError.Cause.(*pq.Error)
		if ok {
			httpError.Detail = pqErr.Detail
		}
		log.Println("client error")
		body, err := httpError.ResponseBody()
		if err != nil {
			log.Printf("An error accured: %v", err)
			w.WriteHeader(500)
		}
		status, headers := httpError.ResponseHeaders()
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
		return
	case ServerError:
		log.Println("server error")
		pqErr := err.(*pq.Error)
		log.Println(pqErr.Code)
		w.WriteHeader(500)
		return
	}

	log.Println(err)
}

type ErrorMissingField string

func (e ErrorMissingField) Error() string {
	return string(e) + " is required."
}

// type ClientError interface {
// 	Error() string
// 	ResponseBody() ([]byte, error)
// 	ResponseHeaders() (int, map[string]string)
// }

type ErrorType int

const (
	ClientError ErrorType = iota
	ServerError
)

type HTTPError struct {
	Type   ErrorType `json:"-"`
	Cause  error     `json:"-"`
	Detail string    `json:"detail"`
	Status int       `json:"-"`
}

func (e HTTPError) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
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

func NewClientError(err error, status int, detail string) error {
	return &HTTPError{
		Type:   ClientError,
		Cause:  err,
		Detail: detail,
		Status: status,
	}
}

func NewServerError(err error, status int, detail string) error {
	return &HTTPError{
		Type:   ClientError,
		Cause:  err,
		Detail: detail,
		Status: status,
	}
}
