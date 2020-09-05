package handlers

import (
	db "PamQ/database"
	"database/sql"
	"errors"
	"fmt"
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

func (user *User) addToDb() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO userinfo VALUES ($1,$2,$3)", user.Username, user.Email, user.HashedPassword); err != nil {
		return fmt.Errorf("User not created. (%s)", err)
	}
	return nil

}

func (u *NewUser) createUser() (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SomeSaltHereMaybeThere"+u.Password), 8)
	if err != nil {
		return nil, fmt.Errorf("Error %s", err)
	}

	user := &User{
		Username:       u.Username,
		Email:          u.Email,
		HashedPassword: string(hashedPassword)}

	err = user.addToDb()
	return user, err
}

func getUserPass(username string) (string, error) {
	var hashedPass string
	db := db.DB
	row := db.QueryRow(`SELECT password FROM userinfo WHERE username=$1`, username)
	err := row.Scan(&hashedPass)
	if err == sql.ErrNoRows {
		return hashedPass, NewHTTPError(err, http.StatusUnauthorized, "Username not found.")
	} else if err != nil {
		return hashedPass, err
	}
	return hashedPass, nil
}
