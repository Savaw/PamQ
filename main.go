package main

import (
	"PamQ/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	startServer()
}

func startServer() {
	// db.InitDB()
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.Handle("/signup", handlers.RootHandler(handlers.SignupHandler)).Methods(http.MethodPost)
	api.Handle("/login", handlers.RootHandler(handlers.LoginHandler)).Methods(http.MethodPost)
	api.Handle("/logout", handlers.RootHandler(handlers.LogoutHandler)).Methods(http.MethodPost)

	quiz := api.PathPrefix("/quiz").Subrouter()
	quiz.Handle("/create", handlers.RootHandler(handlers.CreateQuizHandler)).Methods(http.MethodPost)
	quiz.Handle("/all", handlers.RootHandler(handlers.ListOfQuizesHandler)).Methods(http.MethodGet)
	quiz.Handle("/results", handlers.RootHandler(handlers.QuizResultsHandler)).Methods(http.MethodGet, http.MethodGet)
	quiz.Handle("/{quizID}", handlers.RootHandler(handlers.QuizHandler)).Methods(http.MethodGet, http.MethodPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
