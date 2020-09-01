package handlers

import (
	db "PamQ/database"
	"PamQ/sessions"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username       string `db:"username"`
	Email          string `db:"email"`
	HashedPassword string `db:"password"`
}

type NewUser struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
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
	if matched, err := regexp.Match(`[\w.]+@\w+\.\w+`, []byte(u.Email)); err != nil || matched == false {
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

func (user *User) addUserToDb() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO userinfo VALUES ($1,$2,$3)", user.Username, user.Email, user.HashedPassword); err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("User not created. (%s)", err))
	}
	return nil

}

func (u *NewUser) createUser() (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SomeSaltHereMaybeThere"+u.Password), 8)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error %s", err))
	}

	user := &User{
		Username:       u.Username,
		Email:          u.Email,
		HashedPassword: string(hashedPassword)}

	err = user.addUserToDb()
	return user, err
}

func SignupPostHandler(w http.ResponseWriter, r *http.Request) {

	if sessions.IsLoggedIn(r) {
		w.WriteHeader(http.StatusForbidden)
		returnErrorAsJson(w, "Please logout first.")
		return
	}
	var newUser NewUser

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newUser); err != nil {
		log.Print(err)
		// fmt.Fprintf(w,`{"error": "Error decoding json. (%s)"}`, err )

		returnErrorAsJson(w, fmt.Sprintf(`Error decoding JSON. (%s)`, err))
		return
	}

	if err := newUser.validate(); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
		return
	}

	var user *User
	var err error
	if user, err = newUser.createUser(); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	returnMessageAsJson(w, fmt.Sprintf(`User %s created.`, user.Username))
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) {

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

	if err := bcrypt.CompareHashAndPassword([]byte(storedCred.HashedPassword), []byte("SomeSaltHereMaybeThere"+userCred.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		returnErrorAsJson(w, "Username and password doesn't match.")
		return
	}
	log.Println("success")

	if err := sessions.Login(w, r, userCred.Username); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
		return
	}

	returnMessageAsJson(w, "Login successful.")
}

//LogoutPostHandler handle logout with method post
func LogoutPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := sessions.Logout(w, r); err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
		return
	}
	returnMessageAsJson(w, "Logout successful.")
}
