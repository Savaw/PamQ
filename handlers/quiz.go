package handlers

import (
	db "PamQ/database"
	"PamQ/sessions"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/mitchellh/mapstructure"
)

type Question struct {
	Id           int
	QuizId       int `db:"quiz_id" json:"-"`
	QuestionType int `db:"type" json:"type"` //type=1: multiple choice 	type=2: short answer
	Statement    string
	Option1      string
	Option2      string
	Option3      string
	Option4      string
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

func (q Question) validate() error {
	if len(q.Statement) == 0 {
		return ErrorMissingField("Statement")
	}
	if q.QuestionType == 1 {
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

	for _, value := range q.Questions {
		// fmt.Println(key, value.(map[string]interface{}))
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
		mapstructure.Decode(q, &question)
		question.QuestionType = questionType

		if err := question.validate(); err != nil {
			return quiz, err
		}

		quiz.Questions = append(quiz.Questions, question)
	}

	return quiz, nil
}

func (q Question) addQuestionToDB() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO question (quiz_id, type, statement, option1, option2, option3, option4, answer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", q.QuizId, q.QuestionType, q.Statement, q.Option1, q.Option2, q.Option3, q.Option4, q.Answer); err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("Question not created. (%s)", err))
	}
	return nil
}
func (q *Quiz) addQuizToDB() error {
	db := db.DB

	var quizId int
	row := db.QueryRow("INSERT INTO quiz (creator, name) VALUES ($1, $2) RETURNING id", q.Creator, q.Name)
	err := row.Scan(&quizId)
	if err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("Quiz not created. (%s)", err))
	}

	for _, question := range q.Questions {
		question.QuizId = quizId
		question.addQuestionToDB()

	}
	return nil
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
		return
	}

	var ok bool
	quiz.Creator, ok = sessions.GetUsername(r).(string)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		returnErrorAsJson(w, "Error getting username")
		return
	}

	err = quiz.addQuizToDB()
	if err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	returnMessageAsJson(w, fmt.Sprintf(`Quiz created.`))
}
