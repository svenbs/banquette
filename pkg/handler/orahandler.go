package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/svenbs/banquette/pkg/databases/oracle"
	"github.com/svenbs/banquette/pkg/models"
)

// OracleMethodRouter routes requests via method to the correct handler
func (env *Env) OracleMethodRouter(w http.ResponseWriter, req *http.Request) {
	data := &models.Database{}
	if err := decodeBody(req, data); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusBadRequest, "malformed request")
		return
	}

	if len(data.Token) <= 0 {
		respondErr(w, req, http.StatusBadRequest, "missing token")
		return
	}

	oradb, err := oracle.NewDB(data.Token, env.db)
	if err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusInternalServerError, "internal server error")
		return
	}
	defer oradb.Close()

	switch req.Method {
	case "POST":
		env.createUser(w, req, oradb, data)
	case "DELETE":
		env.dropUser(w, req, oradb, data)
	}
}

// createUser creates a user in a registered database associated to its token
func (env *Env) createUser(w http.ResponseWriter, req *http.Request, oradb oracle.OraDB, data *models.Database) {
	if err := notEmpty(map[string]string{
		"username": data.Username,
		"password": data.Password,
	}); err != nil {
		respondErr(w, req, http.StatusBadRequest, err)
		return
	}

	if err := oradb.CreateUser(data.Username, data.Password); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusInternalServerError, err)
		return
	}

	if err := env.db.BookmarkUser(data.Token, data.Username); err != nil {
		log.Println(err)
		oradb.DropUser(data.Username)
		respondErr(w, req, http.StatusInternalServerError, "could not bookmark user")
		return
	}

	respondMessage(w, req, http.StatusCreated, fmt.Sprintf("user %v created", data.Username))
}

// dropUser drops a user associated to its token.
func (env *Env) dropUser(w http.ResponseWriter, req *http.Request, oradb oracle.OraDB, data *models.Database) {
	if err := notEmpty(map[string]string{
		"username": data.Username,
	}); err != nil {
		respondErr(w, req, http.StatusBadRequest, err)
		return
	}

	if err := oradb.DropUser(data.Username); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := env.db.UnBookmarkUser(data.Token, data.Username); err != nil {
		log.Println(err)
		respondErr(w, req, http.StatusOK, data.Username+" deleted, but could not unbookmark it")
		return
	}

	respondMessage(w, req, http.StatusOK, fmt.Sprintf("user %v removed", data.Username))
}
