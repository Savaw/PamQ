package handlers

import (
	"encoding/json"
	"fmt"
)

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
