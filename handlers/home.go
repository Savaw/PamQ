package handlers

import (
	"PamQ/sessions"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	username, ok := sessions.GetUsername(r)
	if !ok {
		returnMessageAsJson(w, "Welcome! Please login.")
		return
	}
	returnMessageAsJson(w, fmt.Sprintf("Welcome %s!", username))
}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	returnMessageAsJson(w, fmt.Sprintf("In process..."))

}
func returnMessageAsJson(w http.ResponseWriter, msg string) {
	mp := map[string]interface{}{"message": msg}
	js, err := json.Marshal(mp)
	if err != nil {
		log.Printf("err return msg as js")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
