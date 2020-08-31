package handlers

import (
	"encoding/json"
	"net/http"
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
