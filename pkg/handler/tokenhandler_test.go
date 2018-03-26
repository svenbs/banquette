package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/svenbs/banquette/pkg/models"
)

func TestEnv_TokenMethodRouter(t *testing.T) {
	type fields struct {
		db models.Datastore
	}
	tt := []struct {
		name        string
		jsonrequest map[string]string
		value       string
		status      int
		method      string
		msg         string
		err         string
	}{
		// register
		{name: "register internal server error", method: "POST", jsonrequest: map[string]string{"token": "internal", "username": "user", "password": "pass", "dbaddr": "addr", "dbname": "name"}, status: http.StatusInternalServerError, err: "simulated internal server error"},
		{name: "register no values", method: "POST", err: "%"},
		{name: "register no request body", method: "POST", err: "malformed request body"},
		{name: "register no jsonrequest body", method: "POST", err: "malformed request body"},
		{name: "register missing username", method: "POST", jsonrequest: map[string]string{"token": "token", "password": "pass", "dbaddr": "addr", "dbname": "name"}, err: "username is missing"},
		{name: "register missing password", method: "POST", jsonrequest: map[string]string{"token": "token", "username": "user", "dbaddr": "addr", "dbname": "name"}, err: "password is missing"},
		{name: "register missing dbaddr", method: "POST", jsonrequest: map[string]string{"token": "token", "username": "user", "password": "pass", "dbname": "name"}, err: "dbaddr is missing"},
		{name: "register missing dbname", method: "POST", jsonrequest: map[string]string{"token": "token", "username": "user", "password": "pass", "dbaddr": "addr"}, err: "dbname is missing"},
		{name: "register successfull", method: "POST", status: http.StatusCreated, jsonrequest: map[string]string{"token": "token", "username": "user", "password": "pass", "dbaddr": "addr", "dbname": "name"}, msg: "{\"token\":\"sha256token\"}"},
		// update
		{name: "update internal server error", method: "PATCH", jsonrequest: map[string]string{"token": "internal", "username": "user", "password": "pass", "dbaddr": "addr", "dbname": "name"}, status: http.StatusInternalServerError, err: "simulated internal server error"},
		{name: "update no values", method: "PATCH", err: "%"},
		{name: "update no request body", method: "PATCH", err: "malformed request body"},
		{name: "update no jsonrequest body", method: "PATCH", err: "malformed request body"},
		{name: "update missing token", method: "PATCH", jsonrequest: map[string]string{"username": "user", "password": "pass", "dbaddr": "addr", "dbname": "name"}, err: "token is missing"},
		{name: "update missing username", method: "PATCH", jsonrequest: map[string]string{"token": "token", "password": "pass", "dbaddr": "addr", "dbname": "name"}, err: "username is missing"},
		{name: "update missing password", method: "PATCH", jsonrequest: map[string]string{"token": "token", "username": "user", "dbaddr": "addr", "dbname": "name"}, err: "password is missing"},
		{name: "update missing dbaddr", method: "PATCH", jsonrequest: map[string]string{"token": "token", "username": "user", "password": "pass", "dbname": "name"}, err: "dbaddr is missing"},
		{name: "update missing dbname", method: "PATCH", jsonrequest: map[string]string{"token": "token", "username": "user", "password": "pass", "dbaddr": "addr"}, err: "dbname is missing"},
		{name: "update successfull", method: "PATCH", status: http.StatusOK, jsonrequest: map[string]string{"token": "token", "username": "user", "password": "pass", "dbaddr": "addr", "dbname": "name"}, msg: "{\"message\":\"token updated\"}"},
		// delete
		{name: "delete internal server error", method: "DELETE", jsonrequest: map[string]string{"token": "internal", "username": "user", "password": "pass", "dbaddr": "addr", "dbname": "name"}, status: http.StatusInternalServerError, err: "simulated internal server error"},
		{name: "delete no values", method: "DELETE", err: "%"},
		{name: "delete no request body", method: "DELETE", err: "malformed request body"},
		{name: "delete no jsonrequest body", method: "DELETE", value: "non jsonrequest body", err: "malformed request body"},
		{name: "delete missing token", method: "DELETE", jsonrequest: map[string]string{}, err: "token is missing"},
		{name: "delete successfull", method: "DELETE", status: http.StatusOK, jsonrequest: map[string]string{"token": "unregister"}, msg: "{\"message\":\"token deleted\"}"},
	}

	var db *mockDB
	env := &Env{db}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			wr := bufio.NewWriter(&buf)
			r := bufio.NewReader(&buf)
			reqbody := bufio.NewReadWriter(r, wr)

			if tc.jsonrequest != nil {
				fmt.Fprint(reqbody, "{")
				var s string
				for key, val := range tc.jsonrequest {
					s = fmt.Sprintf("%v%q:%q,", s, key, val)
				}
				fmt.Fprint(reqbody, strings.TrimSuffix(s, ","))
				fmt.Fprint(reqbody, "}")
				reqbody.Flush()
			} else {
				fmt.Fprint(reqbody, tc.value)
			}

			req, err := http.NewRequest(tc.method, "localhost:8080/api/v1/oracle", reqbody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}
			rec := httptest.NewRecorder()
			env.TokenMethodRouter(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			expectedStatus := http.StatusBadRequest
			if tc.status > 0 {
				expectedStatus = tc.status
			}

			if tc.err != "" {
				if res.StatusCode != expectedStatus {
					t.Errorf("expected status %v; got %v", expectedStatus, res.StatusCode)
				}
				if tc.err == "%" {
					return
				}
				ferr := fmt.Sprintf("{\"error\":{\"message\":%q}}", tc.err)
				if msg := string(bytes.TrimSpace(b)); msg != ferr {
					t.Errorf("expected message %q; got %q", ferr, msg)
				}
				return
			}

			if msg := string(bytes.TrimSpace(b)); msg != tc.msg {
				t.Errorf("expected message %q; got %q", tc.msg, msg)
			}

			if res.StatusCode != expectedStatus {
				t.Errorf("expected status %v; got %v", expectedStatus, res.Status)
			}
		})
	}
}

type mockDB struct{}

func (db *mockDB) Close() {}

func (db *mockDB) Get(string) (*models.Database, error) {
	return nil, nil
}

func (db *mockDB) RegisterDatabase(data *models.Database) error {
	data.Type = "oracle"

	if data.Token == "internal" {
		return fmt.Errorf("simulated internal server error")
	}

	data.Token = "sha256token"
	return nil
}

// UpdateDatabase updates database credentials for a token
func (db *mockDB) UpdateDatabase(data *models.Database) error {
	if data.Token == "internal" {
		return fmt.Errorf("simulated internal server error")
	}
	data.Token = "sha256token"
	return nil
}

func (db *mockDB) UnregisterDatabase(data *models.Database) error {
	switch data.Token {
	case "unregister":
		return nil
	case "internal":
		return fmt.Errorf("simulated internal server error")
	default:
		return fmt.Errorf("unexpected test token recieved: %v", data.Token)
	}
}

func (db *mockDB) BookmarkUser(token, username string) error {
	if username == "fail_bookmark" {
		return fmt.Errorf("failed to bookmark user %v", username)
	}
	return nil
}

func (db *mockDB) UnBookmarkUser(token, username string) error {
	if username == "fail_unbookmark" {
		return fmt.Errorf("%v deleted, but could not unbookmark it", username)
	}
	return nil
}
