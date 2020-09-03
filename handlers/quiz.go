package handlers

import (
	db "PamQ/database"
	"PamQ/sessions"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

type Question struct {
	Id        int          `json:"id"`
	QuizId    int          `db:"quiz_id" json:"-"`
	QType     QuestionType `db:"type" json:"type"` //type=1: multiple choice 	type=2: short answer
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
	Id       int
	QuizId   int `db:"quiz_id"`
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

func (q Question) check(userAnswer string) AnswerResult {
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

func (q Question) validate() error {
	if len(q.Statement) == 0 {
		return ErrorMissingField("Statement")
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

func (p QuizParticipation) addToDB() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO quiz_participation (quiz_id, username, result) VALUES($1, $2, $3)", p.QuizId, p.Username, p.Result); err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("Result not saved. (%s)", err))
	}
	return nil
}

func (q Question) addToDB() error {
	db := db.DB
	if _, err := db.Query("INSERT INTO question (quiz_id, type, statement, option1, option2, option3, option4, answer) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", q.QuizId, q.QType, q.Statement, q.Option1, q.Option2, q.Option3, q.Option4, q.Answer); err != nil {
		log.Println(err)
		return errors.New(fmt.Sprintf("Question not created. (%s)", err))
	}
	return nil
}
func (q *Quiz) addToDB() (int, error) {
	db := db.DB

	var quizId int
	row := db.QueryRow("INSERT INTO quiz (creator, name,  grading_type, pass_fail, passing_score, not_fail_text,fail_text, allowed_participations) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", q.Creator, q.Name, q.GradingType, q.PassFail, q.PassingScore, q.NotFailText, q.FailText, q.AllowedParticipations)
	err := row.Scan(&quizId)
	if err != nil {
		log.Println(err)
		return quizId, errors.New(fmt.Sprintf("Quiz not created. (%s)", err))
	}

	for _, question := range q.Questions {
		question.QuizId = quizId
		err := question.addToDB()
		if err != nil {
			return quizId, nil
		}

	}
	return quizId, nil
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

	quizID, err := quiz.addToDB()
	if err != nil {
		returnErrorAsJson(w, fmt.Sprintf("%s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	mp := map[string]interface{}{"message": "Quiz created.", "id": quizID}
	js, err2 := json.Marshal(mp)

	if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)

}

func getQuizIdParam(r *http.Request) (int, error) {
	pathParams := mux.Vars(r)

	quizID_, ok := pathParams["quizID"]
	if !ok {
		return 0, errors.New("param not found")
	}

	quizID, err := strconv.Atoi(quizID_)
	if err != nil {
		return 0, errors.New("Page not found")
	}
	return quizID, nil

}

func QuizHandler(w http.ResponseWriter, r *http.Request) {
	quizID, err := getQuizIdParam(r)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var quiz Quiz

	db := db.DB
	err = db.QueryRow(`SELECT * FROM quiz WHERE id=$1`, quizID).Scan(&quiz.Id, &quiz.Creator, &quiz.Name, &quiz.GradingType, &quiz.PassFail, &quiz.PassingScore, &quiz.NotFailText, &quiz.FailText, &quiz.AllowedParticipations)

	if err != nil {
		if err == sql.ErrNoRows {
			returnErrorAsJson(w, "Quiz not found")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)

		log.Fatal(err)
		return
	}

	rows, err := db.Query(`SELECT * FROM question WHERE quiz_id=$1`, quizID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	for rows.Next() {
		var r Question
		err = rows.Scan(&r.Id, &r.QuizId, &r.QType, &r.Statement, &r.Option1, &r.Option2, &r.Option3, &r.Option4, &r.Answer)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		quiz.Questions = append(quiz.Questions, r)
	}

	if r.Method == http.MethodGet {
		for i := range quiz.Questions {
			quiz.Questions[i].Answer = ""
		}

		quiz.FailText = ""
		quiz.NotFailText = ""

		js, err := json.Marshal(&quiz)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	if r.Method == http.MethodPost {
		if !sessions.IsLoggedIn(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// var userAnswers UserAnswers
		var userAnswers map[string]string
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&userAnswers); err != nil {
			returnErrorAsJson(w, fmt.Sprintf(`Error decoding JSON. (%s)`, err))
			return
		}

		mark := 0.0
		totalScore := 0.0
		stats := [4]int{0, 0, 0, 0}
		for _, question := range quiz.Questions {
			userAnswer := userAnswers[strconv.Itoa(question.Id)]
			res := question.check(userAnswer)
			stats[res] += 1
			mark += res.Mark(quiz.GradingType)
			if res != QuestionAnswerNotProvided {
				totalScore += 1
			}
		}

		username, ok := sessions.GetUsername(r).(string)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			returnErrorAsJson(w, "Error getting username")
			return
		}

		participation := QuizParticipation{
			QuizId:   quizID,
			Username: username,
			Score:    mark / totalScore * 100}

		if quiz.PassFail && participation.Score < quiz.PassingScore {
			participation.PassFail = false
			participation.Result = quiz.FailText
		} else {
			participation.PassFail = true
			participation.Result = quiz.NotFailText
		}

		err = participation.addToDB()

		if err != nil {
			returnErrorAsJson(w, fmt.Sprintf("%s", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		mp := map[string]interface{}{"message": "result saved.", "result": participation.Result, "score": participation.Score, "pass": participation.PassFail}
		for i := 0; i < 4; i++ {
			mp[AnswerResult(i).String()] = stats[AnswerResult(i)]
		}

		js, err2 := json.Marshal(mp)

		if err2 != nil {
			http.Error(w, err2.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(js)
	}
}

func ListOfQuizesHandler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	username := query.Get("createdby")

	db := db.DB
	var rows *sql.Rows
	var err error

	dbQuery := `SELECT id, creator, name, grading_type, pass_fail, passing_score, allowed_participations FROM quiz`
	if len(username) != 0 {
		rows, err = db.Query(dbQuery+` WHERE creator=$1`, username)
	} else {
		rows, err = db.Query(dbQuery)
	}

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	quizes := []Quiz{}

	for rows.Next() {
		var quiz Quiz
		err = rows.Scan(&quiz.Id, &quiz.Creator, &quiz.Name, &quiz.GradingType, &quiz.PassFail, &quiz.PassingScore, &quiz.AllowedParticipations)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		quizes = append(quizes, quiz)
	}

	mp := map[string]interface{}{"quizes": quizes}
	js, err := json.Marshal(mp)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return
}
