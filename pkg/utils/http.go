package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func StatusUnauthorized(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintln(w, "{}")
}
func NotFound(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "{}")
}

func BadRequest(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintln(w, "{}")
}

func InternalServerError(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "{}")
}

func Json(w http.ResponseWriter, request *http.Request, code int, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		InternalServerError(w, request)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(code)
		w.Write(payload)
	}
}
func OkJson(w http.ResponseWriter, request *http.Request, data interface{}) {
	Json(w, request, http.StatusOK, data)
}

func CreatedJson(w http.ResponseWriter, request *http.Request, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		InternalServerError(w, request)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusCreated)
		w.Write(payload)
	}
}
