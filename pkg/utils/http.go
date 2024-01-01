package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func NotFound(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "{}")
}

func BadRequest(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusBadGateway)
	fmt.Fprintln(w, "{}")
}

func InternalServerError(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, "{}")
}

func OkJson(w http.ResponseWriter, request *http.Request, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		InternalServerError(w, request)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}
}
