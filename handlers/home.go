package handlers

import (
	"PamQ/sessions"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	username := sessions.GetUsername(r)
	if username == nil {
		returnMessageAsJson(w, "Welcome! Please login.")
		return
	}
	returnMessageAsJson(w, fmt.Sprintf("Welcome %s!", username))
}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	returnMessageAsJson(w, fmt.Sprintf("In process..."))
	pathParams := mux.Vars(r)

	if val, ok := pathParams["quizID"]; ok {
		returnMessageAsJson(w, val)
	}
}
