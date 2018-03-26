package oracle

import (
	"database/sql"
	"fmt"

	"github.com/svenbs/banquette/pkg/models"

	// oracle connection
	_ "github.com/mattn/go-oci8"
)

// OraDB is an interface to a registered database to create or drop users.
type OraDB interface {
	Close()
	CreateUser(username, password string) error
	DropUser(username string) error
}

// DB is a database handle representing a pool of zero or more underlying connections.
// It's safe for concurrent use by multiple goroutines.
type DB struct {
	*sql.DB
}

// NewDB creates a new DB object
func NewDB(token string, tokenstore models.Datastore) (*DB, error) {
	data, err := tokenstore.Get(token)
	if err != nil {
		return nil, fmt.Errorf("could not get database for token (%v): %v", token, err)
	}

	db, err := sql.Open("oci8", data.Username+":"+data.Password+"@"+data.DBAddr+"/"+data.DBName)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database matching token (%v): %v", token, err)
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Close closes the database, releasing any open resources.
func (db *DB) Close() {
	db.DB.Close()
}

// CreateUser creates a user and a tablespace
func (db *DB) CreateUser(username, password string) error {
	if err := notEmpty(map[string]string{
		"username": username,
		"password": password,
	}); err != nil {
		return err
	}

	tablespace := username
	if _, err := db.Exec("CREATE bigfile tablespace " + tablespace + " datafile size 100M autoextend on next 100M"); err != nil {
		return fmt.Errorf("could not create tablespace (%v): %v", tablespace, err)
	}

	_, err := db.Exec("CREATE user " + username + " profile APPUSERS default tablespace " + tablespace + " identified by " + password + " account unlock quota unlimited on " + tablespace)
	if err != nil {
		if _, err := db.Exec("DROP tablespace " + tablespace); err != nil {
			return fmt.Errorf("could not drop tablespace (%v) after user creation failed: %v", tablespace, err)
		}
		return fmt.Errorf("could not create user (%v): %v", username, err)
	}

	if _, err := db.Exec("GRANT GSB to " + username); err != nil {
		return fmt.Errorf("could not grant role GSB to %v: %v", username, err)
	}
	return nil
}

func (db *DB) checkUser(name string) error {
	var value int
	err := db.QueryRow("SELECT count(*) from dba_users where username=:1", name).Scan(&value)
	if err != nil {
		return fmt.Errorf("could not query tablespace: %v", err)
	}

	if value > 0 {
		return fmt.Errorf("user already exists")
	}
	return nil
}

func (db *DB) checkTablespace(name string) error {
	var value int
	err := db.QueryRow("SELECT count(*) from dba_tablespaces where tablespace_name=:1", name).Scan(&value)
	if err != nil {
		return fmt.Errorf("could not query tablespace: %v", err)
	}

	if value > 0 {
		return fmt.Errorf("tablespace already exists")
	}
	return nil
}

// DropUser drops the user and tablespace matching username
func (db *DB) DropUser(username string) error {
	if _, err := db.Exec("DROP user " + username); err != nil {
		return fmt.Errorf("could not drop user (%v): %v", username, err)
	}
	if _, err := db.Exec("DROP tablespace " + username); err != nil {
		return fmt.Errorf("could not drop tablespace (%v): %v", username, err)
	}
	return nil
}

// notEmpty checks if a string inside a map is empty or not
func notEmpty(args map[string]string) error {
	for key, value := range args {
		if len(value) <= 0 {
			return fmt.Errorf("%v is missing", key)
		}
	}
	return nil
}
