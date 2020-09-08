package handlers_test

import (
	"PamQ/handlers"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignUpHandler(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		input      string
		want       string
		statusCode int
	}{
		{
			name:       "Bad username",
			input:      `{"username": "a green crocodile", "password":"S2525fs_23523", "password_confirm": "S2525fs_23523", "email":"sab@oo.com"}`,
			want:       "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Bad email",
			input:      `{"username": "test_user1", "password":"S2525fs_23523", "password_confirm": "S2525fs_23523", "email":"fo.com"}`,
			want:       "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Weak password",
			input:      `{"username": "test_user2", "password":"1234", "password_confirm": "1234", "email":"sab@oo.com"}`,
			want:       "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Not matching password",
			input:      `{"username": "test_user3", "password":"S_wewrw*23", "password_confirm": "S_wewrw*2", "email":"sab@oo.com"}`,
			want:       "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Normal user",
			input:      `{"username": "test_user4", "password":"sY43_h*w", "password_confirm": "sY43_h*w", "email":"sab4@foo.com"}`,
			want:       "",
			statusCode: http.StatusCreated,
		},
		{
			name:       "Duplicate user",
			input:      `{"username": "test_user4", "password":"S_wewrw*23", "password_confirm": "S_wewrw*2", "email":"sab5@oo.com"}`,
			want:       "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Duplicate email",
			input:      `{"username": "test_user5", "password":"S_wewrw*23", "password_confirm": "S_wewrw*2", "email":"sab4@foo.com"}`,
			want:       "",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var js = []byte(tc.input)
			request := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(js))

			request.Header.Set("Content-Type", "application/json")
			responseRecorder := httptest.NewRecorder()
			handler := handlers.RootHandler(handlers.SignupHandler)
			handler.ServeHTTP(responseRecorder, request)
			// fmt.Println(responseRecorder.Body)
			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'", tc.statusCode, responseRecorder.Code)
			}
		})
	}
}
