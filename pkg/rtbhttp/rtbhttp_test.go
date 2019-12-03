/*
 * Author: Chris Lenz <chris@rtbrick.com>
 * Copyright (c) 2016 - 2019, RtBrick, Inc.
 */

package rtbhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteMessage(t *testing.T) {
	type args struct {
		statuscode int
		message    string
	}
	tests := []struct {
		name string
		args args
		body string
	}{
		{"1", args{http.StatusInternalServerError, "test"}, `{"message":"test"}`},
		{"2", args{http.StatusNotFound, "not found"}, `{"message":"not found"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteMessage(rr, tt.args.statuscode, tt.args.message)
			resp := rr.Result()
			if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
				t.Error("Content-Type should be application/json")
			}
			if status := resp.StatusCode; status != tt.args.statuscode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					resp.StatusCode, tt.args.statuscode)
			}
			body, _ := ioutil.ReadAll(resp.Body)
			stringbody := strings.TrimSpace(string(body))
			if stringbody != tt.body {
				t.Errorf("handler returned unexpected body: got %v want %v",
					stringbody, tt.body)
			}
		})
	}
}

func TestWriteAsJSON(t *testing.T) {
	type args struct {
		statuscode int
		message    interface{}
	}
	tests := []struct {
		name string
		args args
		body string
	}{
		{"1", args{http.StatusInternalServerError, "test"}, `"test"`},
		{"2", args{http.StatusNotFound, &Message{Message: "not found"}}, `{"message":"not found"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteAsJSON(rr, tt.args.statuscode, tt.args.message)
			resp := rr.Result()
			if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
				t.Error("Content-Type should be application/json")
			}
			if status := resp.StatusCode; status != tt.args.statuscode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					resp.StatusCode, tt.args.statuscode)
			}
			body, _ := ioutil.ReadAll(resp.Body)
			stringbody := strings.TrimSpace(string(body))
			if stringbody != tt.body {
				t.Errorf("handler returned unexpected body: got %v want %v",
					stringbody, tt.body)
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		message string
	}{
		{"1", `{"message":"not found"}`, "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/projects", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatal(err)
			}
			message := &Message{}
			_ = ReadJSON(req, message)
			if message.Message != tt.message {
				t.Errorf("Wrong Message: got '%v' want '%v'",
					message.Message, tt.message)
			}
		})
	}
}
