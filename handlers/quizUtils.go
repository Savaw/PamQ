package handlers

import (
	db "PamQ/database"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

type Question struct {
	Id        int          `json:"id"`
	QuizID    int          `db:"quiz_id" json:"-"`
	QType     QuestionType `db:"type" json:"type"`
	Statement string       `json:"statement"`
	Option1   string       `json:"option1,omitempty"`
	Option2   string       `json:"option2,omitempty"`
	Option3   string       `json:"option3,omitempty"`
	Option4   string       `json:"option4,omitempty"`
	Answer    string       `json:"answer,omitempty"`
}

type QuestionType int

const (
	MultiChoice = iota + 1
	ShortAnswer
)

type Quiz struct {
	Id                    int        `json:"id"`
	Creator               string     `json:"creator"`
	Name                  string     `json:"name"`
	Questions             []Question `json:"questions,omitempty"`
	GradingType           Grading    `json:"grading_type" db:"grading_type"`
	PassFail              bool       `json:"pass_fail" db:"pass_fail"`
	PassingScore          float64    `json:"passing_score" db:"passing_score"`
	NotFailText           string     `json:"not_fail_text" db:"not_fail_text"`
	FailText              string     `json:"fail_text" db:"fail_text"`
	AllowedParticipations int        `json:"allowed_participation" db:"allowed_participation"`
}

type NewQuiz struct {
	Name                  string        `db,json:"name"`
	NewQuestions          []interface{} `json:"questions"`
	GradingType           Grading       `json:"grading_type" db:"grading_type"`
	PassFail              bool          `json:"pass_fail" db:"pass_fail"`
	PassingScore          float64       `json:"passing_score" db:"passing_score"`
	NotFailText           string        `json:"not_fail_text" db:"not_fail_text"`
	FailText              string        `json:"fail_text" db:"fail_text"`
	AllowedParticipations int           `json:"allowed_participation" db:"allowed_participation"`
}

type QuizParticipation struct {
	ID       int
	QuizID   int `db:"quiz_id"`
	Username string
	Result   string
	Score    float64
	PassFail bool `db:"pass_fail"`
}

// type UserAnswer struct {
// 	QuestionId int `json:"id"`
// 	Answer     string
// }

// type UserAnswers struct {
// 	Answers []UserAnswer
// }

type AnswerResult int

const (
	Wrong AnswerResult = iota
	NoAnswer
	Correct
	QuestionAnswerNotProvided
)

func (a AnswerResult) String() string {
	l := [...]string{"Wrong", "NoAnswer", "Correct", "QuestionAnswerNotProvided"}
	if a >= 0 && a < 4 {
		return l[a]
	}
	return "Unknown"
}

type Grading int

const (
	OnlyCorrect Grading = iota + 1
	WithNegetiveMark
)

func (a AnswerResult) Mark(g Grading) float64 {
	switch g {
	case OnlyCorrect:
		switch a {
		case Correct:
			return 1
		case Wrong, NoAnswer, QuestionAnswerNotProvided:
			return 0
		}
	case WithNegetiveMark:
		switch a {
		case Correct:
			return 1
		case Wrong:
			return -0.25
		case NoAnswer, QuestionAnswerNotProvided:
			return 0
		}
	}

	return 0
}

func (q *Question) check(userAnswer string) AnswerResult {
	uAns := strings.TrimSpace(userAnswer)
	ans := strings.TrimSpace(q.Answer)

	if len(ans) == 0 {
		return QuestionAnswerNotProvided
	}
	if len(uAns) == 0 {
		return NoAnswer
	}

	if strings.EqualFold(uAns, ans) {
		return Correct
	} else {
		return Wrong
	}
}

func (q *Question) validate() error {
	if len(q.Statement) == 0 {
		return ErrorMissingField("Question statement")
	}
	if q.QType == 1 {
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
		return quiz, ErrorMissingField("name")
	}
	if len(q.NewQuestions) == 0 {
		return quiz, ErrorMissingField("questions")
	}

	if q.GradingType < 1 || q.GradingType > 2 {
		return quiz, errors.New("Please enter a valid type for Grading Type. (1 or 2)")
	}

	if q.AllowedParticipations == 0 {
		return quiz, ErrorMissingField("allowed_participations")
	}

	quiz.Name = q.Name
	err := mapstructure.Decode(q, &quiz)
	if err != nil {
		log.Println(err)
		return quiz, err
	}

	for _, value := range q.NewQuestions {
		qu := value.(map[string]interface{})

		qTypeError := errors.New("Please enter a valid type for question. (1 or 2)")

		t, ok := qu["type"].(float64)
		var questionType int
		if ok {
			questionType = int(t)
		} else {
			t, ok := qu["type"].(string)
			if !ok {
				return quiz, qTypeError
			}
			var err error
			questionType, err = strconv.Atoi(t)
			if err != nil {
				return quiz, qTypeError
			}
		}

		if questionType != 1 && questionType != 2 {
			return quiz, qTypeError
		}

		var question Question
		mapstructure.Decode(qu, &question)
		question.QType = QuestionType(questionType)

		if err := question.validate(); err != nil {
			return quiz, err
		}

		quiz.Questions = append(quiz.Questions, question)
	}

	return quiz, nil
}

func (p *QuizParticipation) addToDB() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO quiz_participation (quiz_id, username, result) VALUES($1, $2, $3)", p.QuizID, p.Username, p.Result); err != nil {
		return NewServerError(err, 500, "Quiz participation not saved in database")
	}
	return nil
}

func (q *Question) addToDB() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO question (quiz_id, type, statement, option1, option2, option3, option4, answer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", q.QuizID, q.QType, q.Statement, q.Option1, q.Option2, q.Option3, q.Option4, q.Answer); err != nil {
		return NewServerError(err, 500, "Question not saved in database")
	}
	return nil
}
func (q *Quiz) addToDB() (int, error) {
	db := db.DB

	var quizId int
	row := db.QueryRow("INSERT INTO quiz (creator, name,  grading_type, pass_fail, passing_score, not_fail_text,fail_text, allowed_participations) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", q.Creator, q.Name, q.GradingType, q.PassFail, q.PassingScore, q.NotFailText, q.FailText, q.AllowedParticipations)
	err := row.Scan(&quizId)
	if err != nil {
		return quizId, NewServerError(err, 500, "Quiz not saved in database")
	}

	for _, question := range q.Questions {
		question.QuizID = quizId
		err := question.addToDB()
		if err != nil {
			return quizId, nil
		}

	}
	return quizId, nil
}

func getQuizIdParam(r *http.Request) (int, error) {
	pathParams := mux.Vars(r)

	quizID_, ok := pathParams["quizID"]
	if !ok {
		return 0, NewServerError(nil, 500, "quizId parameter not found")
	}

	quizID, err := strconv.Atoi(quizID_)
	if err != nil {
		return 0, NewClientError(err, http.StatusNotFound, "Page not found")
	}
	return quizID, nil

}
