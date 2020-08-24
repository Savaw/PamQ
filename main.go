package main

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"	
	"encoding/json"
	"fmt"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Welcome"}`))
}

type User struct {
	Id			int		
	Username 	string	
	Password 	string	
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user User 
	if err := decoder.Decode(&user); err != nil {
		log.Print(err)
	}

	fmt.Println(user)
	
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