package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/svenbs/banquette/pkg/models"
)

func TestEnv_createUser(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		request        string
		wantMsg        string
		wantStatusCode int
	}{
		{name: "missing username", method: "POST", request: "{\"token\":\"testtoken\",\"password\":\"testpw\"}", wantMsg: "{\"error\":{\"message\":\"username is missing\"}}", wantStatusCode: http.StatusBadRequest},
		{name: "missing password", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"testuser\"}", wantMsg: "{\"error\":{\"message\":\"password is missing\"}}", wantStatusCode: http.StatusBadRequest},
		{name: "tablespace exists", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"tablespace_exists\",\"password\":\"testpw\"}", wantMsg: "{\"error\":{\"message\":\"could not create tablespace (tablespace_exists)\"}}", wantStatusCode: http.StatusInternalServerError},
		{name: "username exists", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"username_exists\",\"password\":\"testpw\"}", wantMsg: "{\"error\":{\"message\":\"could not create user (username_exists)\"}}", wantStatusCode: http.StatusInternalServerError},
		{name: "grant does not exist", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"grant_does_not_exist\",\"password\":\"testpw\"}", wantMsg: "{\"error\":{\"message\":\"could not grant role GSB to grant_does_not_exist\"}}", wantStatusCode: http.StatusInternalServerError},
		{name: "fail to bookmark", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"fail_bookmark\",\"password\":\"testpw\"}", wantMsg: "{\"error\":{\"message\":\"could not bookmark user\"}}", wantStatusCode: http.StatusInternalServerError},
		{name: "create user successfully", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"testuser\",\"password\":\"testpw\"}", wantMsg: "{\"message\":\"user testuser created\"}", wantStatusCode: http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			reqbody, err := newReadWriter()
			if err != nil {
				t.Fatalf("could not create request body: %v", err)
			}

			_, err = reqbody.Write([]byte(tt.request))
			if err != nil {
				t.Fatalf("could not write to request body: %v", err)
			}
			reqbody.Flush()

			req, err := http.NewRequest(tt.method, "", reqbody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			var oradb *oraMockDB
			var data models.Database
			if err := decodeBody(req, &data); err != nil {
				t.Fatalf("could not decode request body: %v", err)
			}

			var db *mockDB
			env := &Env{db}
			env.createUser(rec, req, oradb, &data)

			res := rec.Result()
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read from response body: %v", err)
			}

			if res.StatusCode != tt.wantStatusCode {
				t.Fatalf("expected status code %v; got %v", tt.wantStatusCode, res.StatusCode)
			}
			if msg := strings.TrimSpace(string(body)); msg != tt.wantMsg {
				t.Fatalf("expected message %q; got %q", tt.wantMsg, msg)
			}
		})
	}
}

func TestEnv_dropUser(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		request        string
		wantMsg        string
		wantStatusCode int
	}{
		{name: "missing username", method: "POST", request: "{\"token\":\"testtoken\"}", wantMsg: "{\"error\":{\"message\":\"username is missing\"}}", wantStatusCode: http.StatusBadRequest},
		{name: "failed to unbookmark", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"fail_unbookmark\"}", wantMsg: "{\"error\":{\"message\":\"fail_unbookmark deleted, but could not unbookmark it\"}}", wantStatusCode: http.StatusOK},
		{name: "drop user successfully", method: "POST", request: "{\"token\":\"testtoken\",\"username\":\"testuser\"}", wantMsg: "{\"message\":\"user testuser removed\"}", wantStatusCode: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			reqbody, err := newReadWriter()
			if err != nil {
				t.Fatalf("could not create request body: %v", err)
			}

			_, err = reqbody.Write([]byte(tt.request))
			if err != nil {
				t.Fatalf("could not write to request body: %v", err)
			}
			reqbody.Flush()

			req, err := http.NewRequest(tt.method, "", reqbody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			var oradb *oraMockDB
			var data models.Database
			if err := decodeBody(req, &data); err != nil {
				t.Fatalf("could not decode request body: %v", err)
			}

			var db *mockDB
			env := &Env{db}
			env.dropUser(rec, req, oradb, &data)

			res := rec.Result()
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read from response body: %v", err)
			}

			if res.StatusCode != tt.wantStatusCode {
				t.Fatalf("expected status code %v; got %v", tt.wantStatusCode, res.StatusCode)
			}
			if msg := strings.TrimSpace(string(body)); msg != tt.wantMsg {
				t.Fatalf("expected message %q; got %q", tt.wantMsg, msg)
			}
		})
	}
}

type oraMockDB struct{}

func (db *oraMockDB) Close() {}

func (db *oraMockDB) CreateUser(username, password string) error {
	tablespace := username
	if username == "tablespace_exists" {
		return fmt.Errorf("could not create tablespace (%v)", tablespace)
	}

	if username == "username_exists" {
		return fmt.Errorf("could not create user (%v)", username)
	}

	if username == "grant_does_not_exist" {
		return fmt.Errorf("could not grant role GSB to %v", username)
	}
	return nil
}

func (db *oraMockDB) DropUser(username string) error {
	if err := notEmpty(map[string]string{
		"username": username,
	}); err != nil {
		return err
	}

	switch username {
	case "error_dropping_user":
		return fmt.Errorf("could not drop user (%v)", username)
	case "error_dropping_tablespace":
		return fmt.Errorf("could not drop tablespace (%v)", username)
	}
	return nil
}
