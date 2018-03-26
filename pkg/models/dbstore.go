package models

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

var (
	tokenTable    = "tokens"
	bookmarkTable = "bookmarks"

	// databaseSecret is set by NewDB()
	databaseSecret = "setme"
)

// Database contains database connect information
type Database struct {
	Token    string
	Type     string
	DBAddr   string
	DBName   string
	Username string
	Password string
}

// Get database information that belongs to a token
func (db *DB) Get(token string) (*Database, error) {
	var v Database
	err := db.QueryRow("SELECT AES_DECRYPT(password, ?) from "+tokenTable+" where token=?", databaseSecret, token).Scan(&v.Password)
	if err != nil {
		return nil, fmt.Errorf("could not decode password, check your database secret: %v", err)
	}
	err = db.QueryRow("SELECT type, dbaddr, dbname, username from "+tokenTable+" where token=?", token).Scan(&v.Type, &v.DBAddr, &v.DBName, &v.Username)
	if err != nil {
		return nil, fmt.Errorf("could not get token information: %v", err)
	}
	return &v, nil
}

// BookmarkUser bookmarks a database user and link it
// to the database token for which it was created
func (db *DB) BookmarkUser(token, username string) error {
	tokenID, err := db.getTokenID(token)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO "+bookmarkTable+" (token_id, dbname) values (?, ?)", tokenID, username)
	if err != nil {
		return err
	}
	return nil
}

// UnBookmarkUser removes a bookmar for a database user created by BookmarkUser
func (db *DB) UnBookmarkUser(token, username string) error {
	tokenID, err := db.getTokenID(token)
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM "+bookmarkTable+" where token_id=? and dbname=?", tokenID, username)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) getTokenID(token string) (int, error) {
	var tokenID int
	err := db.QueryRow("SELECT id from "+tokenTable+" where token=?", token).Scan(&tokenID)
	if err != nil {
		return 0, fmt.Errorf("could not get tokenID: %v", err)
	}
	return tokenID, nil
}

// RegisterDatabase creates and stores a token and access credentials
// for a database that can be used to create other database users.
func (db *DB) RegisterDatabase(data *Database) error {
	data.Type = "oracle"

	if err := db.checkDB(data.DBAddr, data.DBName); err != nil {
		return err
	}
	var err error
	data.Token, err = generateToken(data.DBAddr, data.DBName)
	if err != nil {
		return fmt.Errorf("could not generate token: %v", err)
	}

	if _, err := db.Exec("INSERT INTO "+tokenTable+" (token, type, dbaddr, dbname, username, password) values (?, ?, ?, ?, ?, AES_ENCRYPT(?, ?))", data.Token, data.Type, data.DBAddr, data.DBName, data.Username, data.Password, databaseSecret); err != nil {
		return fmt.Errorf("could not store token: %v", err)
	}
	return nil
}

func generateToken(args ...interface{}) (string, error) {
	h := sha256.New()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if _, err := h.Write([]byte(fmt.Sprint(args...) + strconv.Itoa(r.Int()))); err != nil {
		return "", fmt.Errorf("error hashing token: %v", err)
	}

	return fmt.Sprintf("%x", string(h.Sum(nil))), nil
}

func (db *DB) checkDB(dbaddr, dbname string) error {
	var count int
	err := db.QueryRow("SELECT count(*) FROM "+tokenTable+" where dbaddr=? and dbname=?", dbaddr, dbname).Scan(&count)
	if err != nil {
		return fmt.Errorf("could not connect token database: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("database token already exists")
	}
	return nil
}

// UpdateDatabase updates database credentials for a token
func (db *DB) UpdateDatabase(data *Database) error {
	_, err := db.Exec("UPDATE "+tokenTable+" set token=?, type=?, dbaddr=?, dbname=?, username=?, password=AES_ENCRYPT(?, ?)", data.Token, data.Type, data.DBAddr, data.DBName, data.Username, data.Password, databaseSecret)
	if err != nil {
		return fmt.Errorf("could not update token: %v", err)
	}
	return nil
}

// UnregisterDatabase removes a token and all it's
// linked BookmarkUsers from the datastore
func (db *DB) UnregisterDatabase(data *Database) error {
	if _, err := db.getTokenID(data.Token); err != nil {
		return err
	}
	_, err := db.Exec("DELETE from "+tokenTable+" where token=?", data.Token)
	if err != nil {
		return fmt.Errorf("could not delete token: %v", err)
	}
	return nil
}
