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
	db.InitDB()
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HomeHandler)

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/signup", handlers.SignupPostHandler).Methods(http.MethodPost)
	api.HandleFunc("/login", handlers.LoginPostHandler).Methods(http.MethodPost)
	api.HandleFunc("/logout", handlers.LogoutPostHandler).Methods(http.MethodPost)

	quiz := api.PathPrefix("/quiz").Subrouter()
	quiz.HandleFunc("/create", handlers.CreateQuizHandler).Methods(http.MethodPost)
	quiz.HandleFunc("/{quizID}", handlers.GetQuizHandler).Methods(http.MethodGet)
	quiz.HandleFunc("/{quizID}/edit", handlers.EmptyHandler)
	quiz.HandleFunc("/{quizID}/result", handlers.EmptyHandler)

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
