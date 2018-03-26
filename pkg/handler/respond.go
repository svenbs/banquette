package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func encodeBody(w http.ResponseWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		encodeBody(w, data)
	}
}

func respondMessage(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	respondJSON(w, status, map[string]interface{}{
		"message": fmt.Sprint(args...),
	})
}

func respondErr(w http.ResponseWriter, r *http.Request, status int, args ...interface{}) {
	respondJSON(w, status, map[string]interface{}{
		"error": map[string]interface{}{
			"message": fmt.Sprint(args...),
		},
	})
}
