package handlers

import (
	"net/http"
)

type question interface {
}

type multipleChoiceQuestion struct {
	quizId       int `db:"quiz_id"`
	questionType int `db:"type"`
	statement    string
	option1      string
	option2      string
	option3      string
	option4      string
	answer       string
}

type shortAnswer struct {
	quizId       int `db:"quiz_id"`
	questionType int `db:"type"`
	statement    string
	answer       string
}

type quiz struct {
	id        int
	creator   string
	name      string
	questions []question
}

func CreateQuizHandler(w http.ResponseWriter, r *http.Request) {

}
