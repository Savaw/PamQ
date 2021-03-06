package handlers

import (
	db "PamQ/database"
	"PamQ/sessions"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func CreateQuizHandler(w http.ResponseWriter, r *http.Request) error {
	if !sessions.IsLoggedIn(r) {
		return NewClientError(nil, http.StatusUnauthorized, "Please login first")
	}

	var newQuiz NewQuiz
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newQuiz); err != nil {
		return NewClientError(err, 400, "Bad request : invalid JSON.")
	}

	quiz, err := newQuiz.validate()
	if err != nil {
		return NewClientError(err, http.StatusBadRequest, fmt.Sprintf("Invalid form data: %s", err.Error()))
	}

	var ok bool
	quiz.Creator, ok = sessions.GetUsername(r)
	if !ok {
		return NewServerError(nil, 500, "Error getting username from session")
	}

	quizID, err := quiz.addToDB()
	if err != nil {
		return err
	}

	mp := map[string]interface{}{"message": "Quiz created.", "id": quizID}
	js, err := json.Marshal(mp)
	if err != nil {
		return NewServerError(err, 500, "Error while parsing response body")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(js)
	return nil
}

func QuizHandler(w http.ResponseWriter, r *http.Request) error {
	quizID, err := getQuizIdParam(r)
	if err != nil {
		return err
	}
	var quiz Quiz

	db := db.DB
	err = db.QueryRow(`SELECT * FROM quiz WHERE id=$1`, quizID).Scan(&quiz.Id, &quiz.Creator, &quiz.Name, &quiz.GradingType, &quiz.PassFail, &quiz.PassingScore, &quiz.NotFailText, &quiz.FailText, &quiz.AllowedParticipations, &quiz.DateCreated)

	if err != nil {
		if err == sql.ErrNoRows {
			return NewClientError(err, http.StatusNotFound, "Quiz not found")
		}
		return NewServerError(err, 500, "Error fetching data from database")
	}

	rows, err := db.Query(`SELECT * FROM question WHERE quiz_id=$1`, quizID)
	if err != nil {
		return NewServerError(err, 500, "Error fetching data from database")
	}

	for rows.Next() {
		var q Question
		err = rows.Scan(&q.Id, &q.QuizID, &q.QType, &q.Statement, &q.Option1, &q.Option2, &q.Option3, &q.Option4, &q.Answer)
		if err != nil {
			return NewServerError(err, 500, "Error fetching data from database")
		}
		quiz.Questions = append(quiz.Questions, q)
	}
	loggedIn := sessions.IsLoggedIn(r)
	availableParticipation := quiz.AllowedParticipations
	if loggedIn {
		// quiz.AllowedParticipations
		username, ok := sessions.GetUsername(r)
		if !ok {
			return NewServerError(nil, 500, "Error getting username from session")
		}
		rows, err := db.Query(`SELECT * FROM quiz_participation WHERE quiz_id=$1 AND username=$2`, quizID, username)
		if err != nil {
			return NewServerError(err, 500, "Error fetching data from database")
		}
		count := 0
		for rows.Next() {
			count++
		}
		availableParticipation -= count
	}

	if r.Method == http.MethodGet {
		for i := range quiz.Questions {
			quiz.Questions[i].Answer = ""
		}

		quiz.FailText = ""
		quiz.NotFailText = ""

		comb := struct {
			Quiz
			AvailableParicipation int `json:"available_participation"`
		}{quiz, availableParticipation}

		js, err := json.Marshal(comb)
		if err != nil {
			return NewServerError(err, 500, "Error while parsing response body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return nil
	}

	if r.Method == http.MethodPost {
		if !loggedIn {
			return NewClientError(nil, http.StatusUnauthorized, "Please login first")
		}

		var userAnswers map[string]string
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&userAnswers); err != nil {
			return NewClientError(err, 400, "Bad request : invalid JSON.")
		}

		if availableParticipation <= 0 {
			return NewClientError(nil, http.StatusBadRequest, "Your participation limit for this quiz has been reached")
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

		username, ok := sessions.GetUsername(r)
		if !ok {
			return NewServerError(nil, 500, "Error getting username from session")
		}

		participation := QuizParticipation{
			QuizID:   quizID,
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
			return err
		}

		mp := map[string]interface{}{"message": "result saved.", "result": participation.Result, "score": participation.Score, "pass": participation.PassFail}
		for i := 0; i < 4; i++ {
			mp[AnswerResult(i).String()] = stats[AnswerResult(i)]
		}

		js, err := json.Marshal(mp)
		if err != nil {
			return NewServerError(err, 500, "Error while parsing response body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(js)
		return nil
	}
	return nil
}

func ListOfQuizesHandler(w http.ResponseWriter, r *http.Request) error {

	query := r.URL.Query()
	username := query.Get("createdby")

	db := db.DB
	var rows *sql.Rows
	var err error

	dbQuery := `SELECT id, creator, name, grading_type, pass_fail, passing_score, allowed_participations, date_created FROM quiz`
	if len(username) != 0 {
		rows, err = db.Query(dbQuery+` WHERE creator=$1`, username)
	} else {
		rows, err = db.Query(dbQuery)
	}

	if err != nil {
		return NewServerError(err, 500, "Error fetching data from database")
	}
	quizes := []Quiz{}

	for rows.Next() {
		var quiz Quiz
		err = rows.Scan(&quiz.Id, &quiz.Creator, &quiz.Name, &quiz.GradingType, &quiz.PassFail, &quiz.PassingScore, &quiz.AllowedParticipations, &quiz.DateCreated)
		if err != nil {
			return NewServerError(err, 500, "Error fetching data from database")
		}
		quizes = append(quizes, quiz)
	}

	mp := map[string]interface{}{"quizes": quizes}
	js, err := json.Marshal(mp)
	if err != nil {
		return NewServerError(err, 500, "Error while parsing response body")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}

func QuizResultsHandler(w http.ResponseWriter, r *http.Request) error {
	if !sessions.IsLoggedIn(r) {
		return NewClientError(nil, http.StatusUnauthorized, "Please login first")
	}

	username, ok := sessions.GetUsername(r)
	if !ok {
		return NewServerError(nil, 500, "Error getting username from sessions")
	}

	db := db.DB
	rows, err := db.Query(`SELECT * FROM quiz_participation WHERE username=$1`, username)
	if err != nil {
		return NewServerError(err, 500, "Error fetching data from database")
	}
	var listOfTakenQuiz []QuizParticipation
	for rows.Next() {
		var qp QuizParticipation
		var score sql.NullFloat64
		var passFail sql.NullBool
		err := rows.Scan(&qp.ID, &qp.QuizID, &qp.Username, &qp.Result, &score, &passFail, &qp.DateCreated)
		if score.Valid {
			qp.Score = score.Float64
		}
		if passFail.Valid {
			qp.PassFail = passFail.Bool
		}
		if err != nil {
			return NewServerError(err, 500, "Error fetching data from database")
		}
		listOfTakenQuiz = append(listOfTakenQuiz, qp)
	}

	mp := map[string]interface{}{"participations": listOfTakenQuiz}
	js, err := json.Marshal(mp)
	if err != nil {
		return NewServerError(err, 500, "Error while parsing response body")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}
