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
	Id        int        `json:"id"`
	Creator   string     `json:"creator"`
	Name      string     `json:"name"`
	Questions []Question `json:"questions"`
}

type NewQuiz struct {
	Name      string
	Questions []interface{}
}

type QuizParticipation struct {
	Id       int
	QuizId   int `db:"quiz_id"`
	Username string
	Result   string
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

func (a AnswerResult) Mark() int {
	switch a {
	case Correct:
		return 1
	case Wrong, NoAnswer, QuestionAnswerNotProvided:
		return 0
	}
	return 0
}

func (q Question) check(userAnswer string) AnswerResult {
	if len(q.Answer) == 0 {
		return QuestionAnswerNotProvided
	}
	if len(userAnswer) == 0 {
		return NoAnswer
	}

	if userAnswer == q.Answer {
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
	row := db.QueryRow("INSERT INTO quiz (creator, name) VALUES ($1, $2) RETURNING id", q.Creator, q.Name)
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
	err = db.QueryRow(`SELECT * FROM quiz WHERE id=$1`, quizID).Scan(&quiz.Id, &quiz.Creator, &quiz.Name)
	if err != nil {
		fmt.Println(err)
		if err == sql.ErrNoRows {
			returnErrorAsJson(w, "Quiz not found")
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rows, err := db.Query(`SELECT * FROM question WHERE quiz_id=$1`, quizID)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return

	}

	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM question WHERE quiz_id=$1", quizID)
	err = row.Scan(&count)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		var r Question
		err = rows.Scan(&r.Id, &r.QuizId, &r.QType, &r.Statement, &r.Option1, &r.Option2, &r.Option3, &r.Option4, &r.Answer)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		quiz.Questions = append(quiz.Questions, r)
	}

	if r.Method == http.MethodGet {
		for i := range quiz.Questions {
			quiz.Questions[i].Answer = ""
		}

		js, err := json.Marshal(&quiz)
		if err != nil {
			fmt.Println(err)
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

		fmt.Println(userAnswers)

		mark := 0
		stats := [4]int{0, 0, 0, 0}
		for _, question := range quiz.Questions {
			userAnswer := userAnswers[strconv.Itoa(question.Id)]
			fmt.Println("ans " + userAnswer)
			res := question.check(userAnswer)
			fmt.Println(res)
			stats[res] += 1
			mark += res.Mark()
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
			Result:   strconv.Itoa(mark)}

		err = participation.addToDB()

		if err != nil {
			returnErrorAsJson(w, fmt.Sprintf("%s", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		mp := map[string]interface{}{"message": "result saved.", "result": participation.Result}
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
