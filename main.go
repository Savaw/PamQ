package main

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"	
	"encoding/json"
	"fmt"
	"errors"
	"regexp"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Welcome"}`))
}

type User struct {	
	Username	string	
	Email		string
	Password 	string	
}

type NewUser struct {
	Username		string
	Email			string
	Password		string
	PasswordConfirm string
}

type ErrorMissingField string

func (e ErrorMissingField) Error() string {
	return string(e) + " is required."
}

func (u *NewUser) validate() error {
	if len(u.Username) == 0 {
		return ErrorMissingField("Username")
	}
	if len(u.Email) == 0 {
		return ErrorMissingField("Email")
	}
	if matched, err := regexp.Match(`[\w.]+@\w+.\w+`, []byte(u.Email)); err != nil || matched == false {
		return errors.New("Please enter a valid email.")
	}
	if len(u.Password) == 0 {
		return ErrorMissingField("Password")
	}
	if u.Password != u.PasswordConfirm {
		return errors.New("Password doesn't match.")
	}
	return nil		
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// var user User 
	var newUser NewUser

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newUser); err != nil {
		log.Print(err)
		// fmt.Fprintf(w,`{"error": "Error decoding json. (%s)"}`, err )

		// js, err2 := json.Marshal(fmt.Sprintf(`{"error": "Error decoding json. (%s)"}`, err)); 
		js, err2 := json.Marshal(map [string]string{"error":fmt.Sprintf(`Error decoding JSON. (%s)`, err)}); 

		if err2 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		  }
		w.Write(js)
		return
	}

	if err := newUser.validate(); err != nil {
		js, err2 := json.Marshal(map [string]string{"error": fmt.Sprintf("%s",err)}); 

		if err2 != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		  }
		w.Write(js)
		return
	}

}

func main() {
	r := mux.NewRouter()	
	
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/", homeHandler)
	api.HandleFunc("/signup", signupPostHandler).Methods(http.MethodPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}