package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lib/pq"
)

type RootHandler func(http.ResponseWriter, *http.Request) error

func (fn RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err == nil {
		return
	}
	log.Printf("An error accured: %v %T", err, err)

	httpError, ok := err.(*HTTPError)
	if !ok {
		log.Println("Not http error")
		w.WriteHeader(500)
		return
	}

	switch httpError.Type {
	case ClientError:
		pqErr, ok := httpError.Cause.(*pq.Error)
		if ok {
			httpError.Detail = pqErr.Detail
		}
		log.Println("client error")
		body, err := httpError.ResponseBody()
		if err != nil {
			log.Printf("An error accured: %v", err)
			w.WriteHeader(500)
		}
		status, headers := httpError.ResponseHeaders()
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
		return
	case ServerError:
		log.Println("server error")
		pqErr := err.(*pq.Error)
		log.Println(pqErr.Code)
		w.WriteHeader(500)
		return
	}

	log.Println(err)
}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	mp := map[string]interface{}{"message": "In process.."}
	js, err := json.Marshal(mp)
	if err != nil {
		log.Printf("err return msg as js")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
