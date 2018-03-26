package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/svenbs/banquette/pkg/models"
)

// TokenMethodRouter routes incoming Requests regarding their request method
func (env *Env) TokenMethodRouter(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		env.registerDatabase(w, req)
	case "PATCH":
		env.updateDatabase(w, req)
	case "DELETE":
		env.unregisterDatabase(w, req)
	}
}

// RegisterDatabase registers a database and its credentials
// it returns a token to the user which can be used to
// create users in the registered databse.
func (env *Env) registerDatabase(w http.ResponseWriter, req *http.Request) {
	data := &models.Database{}
	if err := decodeBody(req, data); err != nil {
		respondErr(w, req, http.StatusBadRequest, "malformed request body")
		return
	}

	err := notEmpty(map[string]string{
		"dbaddr":   data.DBAddr,
		"dbname":   data.DBName,
		"username": data.Username,
		"password": data.Password,
	})
	if err != nil {
		respondErr(w, req, http.StatusBadRequest, err)
		return
	}

	if err := env.db.RegisterDatabase(data); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"token": data.Token,
	})
}

// updateDatabase updates a registered database and its credentials.
func (env *Env) updateDatabase(w http.ResponseWriter, req *http.Request) {
	data := &models.Database{}
	if err := decodeBody(req, data); err != nil {
		log.Printf("could not decode request body: %v", err)
		respondErr(w, req, http.StatusBadRequest, "malformed request body")
		return
	}

	err := notEmpty(map[string]string{
		"token":    data.Token,
		"dbaddr":   data.DBAddr,
		"dbname":   data.DBName,
		"username": data.Username,
		"password": data.Password,
	})
	if err != nil {
		respondErr(w, req, http.StatusBadRequest, err)
		return
	}

	if err := env.db.UpdateDatabase(data); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusInternalServerError, err)
		return
	}

	respondMessage(w, req, http.StatusOK, "token updated")
}

// unregisterDatabase unregisters a database with its credentials.
func (env *Env) unregisterDatabase(w http.ResponseWriter, req *http.Request) {
	data := &models.Database{}
	if err := decodeBody(req, data); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusBadRequest, "malformed request body")
		return
	}

	err := notEmpty(map[string]string{
		"token": data.Token,
	})
	if err != nil {
		respondErr(w, req, http.StatusBadRequest, err)
		return
	}

	if err := env.db.UnregisterDatabase(data); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusInternalServerError, err)
		return
	}

	respondMessage(w, req, http.StatusOK, "token deleted")
}

func notEmpty(args map[string]string) error {
	for key, value := range args {
		if len(value) <= 0 {
			return fmt.Errorf("%v is missing", key)
		}
	}
	return nil
}
