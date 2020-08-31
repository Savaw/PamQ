package handlers

import (
	"PamQ/sessions"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func returnAsJson(w http.ResponseWriter, mp map[string]interface{}) {
	js, err2 := json.Marshal(mp)

	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func returnErrorAsJson(w http.ResponseWriter, err string) {
	returnAsJson(w, map[string]interface{}{"error": err})
}

func returnMessageAsJson(w http.ResponseWriter, msg string) {
	returnAsJson(w, map[string]interface{}{"message": msg})
}

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
