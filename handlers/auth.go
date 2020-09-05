package handlers

import (
	"PamQ/sessions"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func SignupPostHandler(w http.ResponseWriter, r *http.Request) error {
	if sessions.IsLoggedIn(r) {
		return NewClientError(nil, http.StatusForbidden, "Logout in order to signup.")
	}

	var newUser NewUser
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newUser); err != nil {
		return NewClientError(err, 400, "Bad request : invalid JSON.")
	}

	if err := newUser.validate(); err != nil {
		return NewClientError(err, http.StatusBadRequest, "Invalid form")
	}

	var user *User
	var err error
	if user, err = newUser.createUser(); err != nil {
		return NewServerError(err, 500, "Create user error")
	}

	w.WriteHeader(http.StatusCreated)
	returnMessageAsJson(w, fmt.Sprintf(`User %s created.`, user.Username))
	return nil
}

func LoginPostHandler(w http.ResponseWriter, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	var userCred NewUser
	if err := decoder.Decode(&userCred); err != nil {
		return NewClientError(err, 400, "Bad request : invalid JSON.")
	}

	hashedPass, err := getUserPass(userCred.Username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte("SomeSaltHereMaybeThere"+userCred.Password)); err != nil {
		return NewClientError(err, http.StatusUnauthorized, "Username and password doesn't match.")
	}

	if err := sessions.Login(w, r, userCred.Username); err != nil {
		return NewServerError(err, 500, "Sessions login error")
	}

	w.Header().Set("Content-Type", "application/json")
	mp := map[string]interface{}{"message": "Login succesful.", "username": userCred.Username}
	js, err := json.Marshal(mp)
	if err != nil {
		return NewServerError(err, 500, "Error while parsing response body")
	}
	w.Write(js)
	return nil
}

func LogoutPostHandler(w http.ResponseWriter, r *http.Request) error {
	if err := sessions.Logout(w, r); err != nil {
		return NewServerError(err, 500, "Sessions logout error")
	}
	returnMessageAsJson(w, "Logout successful.")
	return nil
}
