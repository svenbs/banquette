package handler

import (
	"fmt"

	"github.com/svenbs/banquette/pkg/models"
)

// Env is used to interface with models.Datastore
type Env struct {
	db models.Datastore
}

// InitDB initializes the database to store registered databases
// and bookmark users created by banquette.
func InitDB(driver, secret, dsn string) (*Env, error) {
	db, err := models.NewDB(driver, secret, dsn)
	if err != nil {
		return nil, fmt.Errorf("could not create database connection: %v", err)
	}
	return &Env{db}, nil
}

// Close closes the database, releasing any open resources.
func (env *Env) Close() {
	env.db.Close()
}
