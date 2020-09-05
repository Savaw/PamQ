package main

import (
	db "PamQ/database"

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
	db.InitDB()
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)

	api := r.PathPrefix("/api").Subrouter()
	api.Handle("/signup", handlers.RootHandler(handlers.SignupPostHandler)).Methods(http.MethodPost)
	api.Handle("/login", handlers.RootHandler(handlers.LoginPostHandler)).Methods(http.MethodPost)
	api.Handle("/logout", handlers.RootHandler(handlers.LogoutPostHandler)).Methods(http.MethodPost)

	quiz := api.PathPrefix("/quiz").Subrouter()
	quiz.HandleFunc("/create", handlers.CreateQuizHandler).Methods(http.MethodPost)
	quiz.HandleFunc("/all", handlers.ListOfQuizesHandler).Methods(http.MethodGet)
	quiz.HandleFunc("/{quizID}", handlers.QuizHandler).Methods(http.MethodGet, http.MethodPost)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
