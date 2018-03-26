package models

import (
	"database/sql"

	// mysql driver available
	_ "github.com/go-sql-driver/mysql"
)

// Datastore The Datastore interface wraps the basic methods to
// store registered databases and BookmarkUser users inside those databases.
type Datastore interface {
	Close()
	Get(token string) (*Database, error)
	BookmarkUser(token, username string) error
	UnBookmarkUser(token, username string) error
	RegisterDatabase(data *Database) error
	UpdateDatabase(data *Database) error
	UnregisterDatabase(data *Database) error
}

// DB is a database handle representing a pool of zero or more underlying connections.
// It's safe for concurrent use by multiple goroutines.
type DB struct {
	*sql.DB
}

// NewDB creates a new DB Object
// dataSourceName format ist given by the driver
// possible drivers are: "mysql"
// mysql dsn example: user:password@(dbaddr)/database
func NewDB(driver, secret, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	databaseSecret = secret
	return &DB{db}, nil
}

// Close closes the database, releasing any open resources.
func (db *DB) Close() {
	db.DB.Close()
}
