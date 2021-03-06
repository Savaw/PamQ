package sessions

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("pass")) //TODO use env variable

func IsLoggedIn(r *http.Request) bool {
	session, _ := Store.Get(r, "session")
	if session.Values["loggedin"] == true {
		return true
	}
	return false
}

func Login(w http.ResponseWriter, r *http.Request, username string) error {
	session, err := Store.Get(r, "session")
	if err == nil && !IsLoggedIn(r) {
		session.Values["loggedin"] = true
		session.Values["username"] = username
		session.Save(r, w)
		return nil
	}
	return err
}

func Logout(w http.ResponseWriter, r *http.Request) error {
	session, err := Store.Get(r, "session")
	if err == nil && session.Values["loggedin"] == true {
		session.Values["loggedin"] = false
		session.Save(r, w)
	}
	return err
}

func GetUsername(r *http.Request) (string, bool) {
	session, _ := Store.Get(r, "session")
	if session.Values["loggedin"] == true {
		val := session.Values["username"]
		username, ok := val.(string)
		if !ok {
			return "", false
		}
		return username, true
	}
	return "", false
}
