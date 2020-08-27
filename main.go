package main

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"	
	"encoding/json"
	"fmt"
	"errors"
	"regexp"
	"database/sql"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

const (
	DB_USER = "postgres"
	DB_NAME = "go_db_test"
)

func GetDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable", DB_USER, DB_NAME)

	var err error
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Welcome"}`))
}

type User struct {	
	Username		string	`db:"username"`
	Email			string	`db:"email"`
	HashedPassword 	string	`db:"password"`
}

type NewUser struct {
	Username		string	`json:"username"`
	Email			string	`json:"email"`
	Password		string	`json:"password"`
	PasswordConfirm string	`json:"password_confirm"`
}

type ErrorMissingField string

func (e ErrorMissingField) Error() string {
	return string(e) + " is required."
}

func (u *NewUser) Validate() error {
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
		return errors.New("Passwords don't match.")
	}
	return nil		
}

func (u * NewUser) CreateUser() (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SomeSaltHereMaybeThere" + u.Password), 8)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error %s", err))
	}
	
	user := &User{
		Username: u.Username,
		Email: u.Email,
		HashedPassword: string(hashedPassword)}

	db := GetDB()
	if _, err := db.Query("INSERT INTO userinfo VALUES ($1,$2,$3)", user.Username, user.Email, user.HashedPassword); err != nil {
		log.Println(err)
		return nil, errors.New(fmt.Sprintf("User not created. (%s)", err))
	}

	return user, nil
}

func returnAsJson(w http.ResponseWriter, mp map [string]interface{}) {
	js, err2 := json.Marshal(mp); 

	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	  }
	w.Write(js)
}

func returnErrorAsJson(w http.ResponseWriter, err string) {
	returnAsJson(w, map [string]interface{}{"error":err})
}

func returnMessageAsJson(w http.ResponseWriter, msg string) {
	returnAsJson(w, map [string]interface{}{"message":msg})
}


func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// var user User 
	var newUser NewUser

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newUser); err != nil {
		log.Print(err)
		// fmt.Fprintf(w,`{"error": "Error decoding json. (%s)"}`, err )

		returnErrorAsJson(w, fmt.Sprintf(`Error decoding JSON. (%s)`, err))
		return
	}

	if err := newUser.Validate(); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s",err))
		return
	}

	var user *User
	var err error
	if user, err = newUser.CreateUser(); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s",err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	returnMessageAsJson(w, fmt.Sprintf(`User %s created.`, user.Username))
	return
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