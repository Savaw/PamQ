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
	db "PamQ/database"
	"PamQ/sessions"
)


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

	db := db.DB
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

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	var userCred NewUser
	if err := decoder.Decode(&userCred); err != nil {
		returnErrorAsJson(w, fmt.Sprintf(`Error decoding JSON. (%s)`, err))
		return
	}

	var storedCred User

	db := db.DB
	row := db.QueryRow(`SELECT password FROM userinfo WHERE username=$1`, userCred.Username)
	err := row.Scan(&storedCred.HashedPassword)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized)
		returnErrorAsJson(w, "Username not found.")
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte (storedCred.HashedPassword), []byte ("SomeSaltHereMaybeThere" + userCred.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		returnErrorAsJson(w, "Username and password doesn't match.")
		return
	}
	log.Println("success")

	if err := sessions.Login(w, r, userCred.Username); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s",err))
		return
	}

	http.Redirect(w, r, "/", 302)

	return
}

func logoutPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := sessions.Logout(w, r); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s",err))
		return
	}
	http.Redirect(w, r, "/", 302)
}


func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	username := sessions.GetUsername(r)
	if username == nil {
		returnMessageAsJson(w, "Welcome! Please login.")
		return
	}
	returnMessageAsJson(w, fmt.Sprintf("Welcome %s!", username))
	
}

func main() {
	db.InitDB()
	r := mux.NewRouter()	
	r.HandleFunc("/", homeHandler)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/signup", signupPostHandler).Methods(http.MethodPost)
	api.HandleFunc("/login", loginPostHandler).Methods(http.MethodPost)
	api.HandleFunc("/logout", logoutPostHandler).Methods(http.MethodPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}