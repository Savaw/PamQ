package handlers

import (
	"PamQ/sessions"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Question interface {
	validate() error
}

type MultipleChoiceQuestion struct {
	Id           int
	QuizId       int `db:"quiz_id" json:"-"`
	QuestionType int `db:"type" json:"type"`
	Statement    string
	Option1      string
	Option2      string
	Option3      string
	Option4      string
	Answer       string
}

type ShortAnswer struct {
	Id           int
	QuizId       int `db:"quiz_id" json:"-"`
	QuestionType int `db:"type" json:"type"`
	Statement    string
	Answer       string
}

type Quiz struct {
	Id        int
	Creator   string
	Name      string
	Questions []Question
}

type NewQuiz struct {
	Name      string
	Questions []interface{}
}

func (q ShortAnswer) validate() error {
	if len("statement") == 0 {
		return ErrorMissingField("Statement")
	}
	return nil
}

func (q MultipleChoiceQuestion) validate() error {
	if len(q.Statement) == 0 {
		return ErrorMissingField("Statement")
	}
	if len(q.Option1) == 0 {
		return ErrorMissingField("Option1")
	}
	if len(q.Option2) == 0 {
		return ErrorMissingField("Option2")
	}

	answerError := errors.New("Please enter a valid number as answer for question.")
	if len(q.Answer) != 0 {
		answer, err := strconv.Atoi(q.Answer)
		if err != nil || answer < 1 || answer > 4 {
			return answerError
		}
	}
	return nil
}

func (q *NewQuiz) validate() (Quiz, error) {
	var quiz Quiz
	if len(q.Name) == 0 {
		return quiz, ErrorMissingField("Name")
	}
	if len(q.Questions) == 0 {
		return quiz, ErrorMissingField("Questions")
	}

	quiz.Name = q.Name

	for key, value := range q.Questions {
		fmt.Println(key, value.(map[string]interface{}))
		q := value.(map[string]interface{})

		qTypeError := errors.New("Please enter a valid type for question. (1 or 2)")
		t, ok := q["type"].(string)
		if !ok {
			return quiz, qTypeError
		}
		questionType, err := strconv.Atoi(t)
		if err != nil || (questionType != 1 && questionType != 2) {
			return quiz, qTypeError
		}

		var question Question
		switch questionType {
		case 1:
			var multiChoice MultipleChoiceQuestion
			mapstructure.Decode(q, &multiChoice)
			question = multiChoice
		case 2:
			var shortAns ShortAnswer
			mapstructure.Decode(q, &shortAns)
			question = shortAns
		}

		if err := question.validate(); err != nil {
			return quiz, err
		}

		quiz.Questions = append(quiz.Questions, question)

	}

	return quiz, nil
}

func CreateQuizHandler(w http.ResponseWriter, r *http.Request) {
	if !sessions.IsLoggedIn(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var newQuiz NewQuiz
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newQuiz); err != nil {
		returnErrorAsJson(w, fmt.Sprintf(`Error decoding JSON. (%s)`, err))
		return
	}

	quiz, err := newQuiz.validate()
	if err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
	}
	fmt.Println(quiz)

}
