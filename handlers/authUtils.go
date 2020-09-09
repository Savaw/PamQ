package handlers

import (
	db "PamQ/database"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"unicode"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username       string   `db:"username"`
	Email          string   `db:"email"`
	HashedPassword string   `db:"password"`
	DateCreated    JSONTime `json:"date_created" db:"date_created"`
}

type NewUser struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

func validatePassword(password string) bool {
	digit := false
	letter := false
	for _, c := range password {
		if unicode.IsDigit(c) {
			digit = true
		}
		if unicode.IsLetter(c) {
			letter = true
		}
	}
	if len(password) < 8 || !digit || !letter {
		return false
	}
	return true
}

func (u *NewUser) validate() error {
	if matched, err := regexp.Match(`^[A-Za-z]+[A-Za-z0-9]*(?:[_.][A-Za-z0-9]+)*$`, []byte(u.Username)); err != nil || matched == false || len(u.Username) < 3 || len(u.Username) > 30 {
		return errors.New("Please enter a valid username.")
	}
	if len(u.Email) == 0 {
		return ErrorMissingField("Email")
	}
	if matched, err := regexp.Match(`[\w.]+@\w+\.\w+`, []byte(u.Email)); err != nil || matched == false {
		return errors.New("Please enter a valid email.")
	}
	if !validatePassword(u.Password) {
		return errors.New("Please enter a valid password. (at least 8 characters, one digit and one letter)")
	}
	if u.Password != u.PasswordConfirm {
		return errors.New("Passwords don't match.")
	}
	return nil
}

func (user *User) addToDb() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO userinfo VALUES ($1,$2,$3)", user.Username, user.Email, user.HashedPassword); err != nil {
		return err
	}
	return nil

}

func (u *NewUser) createUser() (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("SomeSaltHereMaybeThere"+u.Password), 8)
	if err != nil {
		return nil, err
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
		return hashedPass, NewClientError(err, http.StatusUnauthorized, "Username not found.")
	} else if err != nil {
		return hashedPass, err
	}
	return hashedPass, nil
}
