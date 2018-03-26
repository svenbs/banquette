package handler

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_decodeBody(t *testing.T) {
	tt := []struct {
		name    string
		json    string
		value   interface{}
		wantErr bool
	}{
		{name: "empty json", json: "{}", value: struct{ key string }{}},
		{name: "empty struct as value", json: "{\"key\":\"value\"}", value: struct{}{}},
		{name: "success decode json", json: "{\"key\":\"value\"}", value: struct{ key string }{}},
		{name: "fail", json: "", value: struct{}{}, wantErr: true},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			reqbody, err := newReadWriter()
			if err != nil {
				t.Fatalf("could not create read writer: %v", err)
			}

			reqbody.Write([]byte(tc.json))
			reqbody.Flush()

			req, err := http.NewRequest("", "", reqbody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			if err := decodeBody(req, &tc.value); (err != nil) != tc.wantErr {
				t.Errorf("decodeBody() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func Test_encodeBody(t *testing.T) {
	tt := []struct {
		name    string
		value   interface{}
		wantMsg string
		wantErr bool
	}{
		{name: "empty struct", value: struct{}{}, wantMsg: "{}"},
		{name: "empty string not encoded", value: "", wantMsg: "\"\""},
		{name: "exported struct", value: struct{ Key string }{Key: "test"}, wantMsg: "{\"Key\":\"test\"}"},
		{name: "unexported struct", value: struct{ key string }{key: "test"}, wantMsg: "{}"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			if err := encodeBody(rec, &tc.value); (err != nil) != tc.wantErr {
				t.Fatalf("encodeBody() error = %v, wantErr %v", err, tc.wantErr)
			}

			body, err := ioutil.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("could not read response body: %v", err)
			}

			got := strings.TrimSpace(string(body))
			if got != tc.wantMsg {
				t.Fatalf("expected %q; got %q", tc.wantMsg, got)
			}
		})
	}
}

func Test_respondJSON(t *testing.T) {
	type args struct {
		status  int
		data    interface{}
		wantMsg string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "200 response msg", args: args{status: 200, data: struct{ Key string }{Key: "test"}, wantMsg: "{\"Key\":\"test\"}"}},
		{name: "201 response complex struct", args: args{status: 201, data: struct {
			Key    string
			Struct interface{}
		}{Key: "test", Struct: struct {
			Key   string
			Value int
		}{Key: "test2", Value: 4}}, wantMsg: "{\"Key\":\"test\",\"Struct\":{\"Key\":\"test2\",\"Value\":4}}"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			respondJSON(rec, tt.args.status, tt.args.data)

			res := rec.Result()
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response body: %v", err)
			}

			if res.StatusCode != tt.args.status {
				t.Fatalf("respondJSON() expected status %v; got %v", tt.args.status, res.StatusCode)
			}
			if msg := strings.TrimSpace(string(body)); msg != tt.args.wantMsg {
				t.Fatalf("respondJSON() expected response %q; got %q", tt.args.wantMsg, msg)
			}
		})
	}
}

func newReadWriter() (*bufio.ReadWriter, error) {
	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)
	r := bufio.NewReader(&buf)

	return bufio.NewReadWriter(r, wr), nil
}
