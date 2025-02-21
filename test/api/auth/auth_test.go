package auth_test

import (
	auth_api "blog/api/auth"
	"blog/config"
	"blog/db/auth"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	err := os.Chdir("../../..")
	if err != nil {
		panic(err)
	}
	config.DBFile = ":memory:"
	err = config.Setup()
	if err != nil {
		panic(err)
	}
	err = config.InitDB()
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestRegister(t *testing.T) {
	tests := []struct {
		user   auth.User
		status int
	}{
		{auth.User{Username: "user", Password: "password"}, http.StatusOK},
		{auth.User{Username: "guest", Password: "password"}, http.StatusOK},
		{auth.User{Username: "user", Password: "password"}, http.StatusConflict},
		{auth.User{Username: "username"}, http.StatusBadRequest},
		{auth.User{Password: "password"}, http.StatusBadRequest},
		{auth.User{}, http.StatusBadRequest},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(test.user)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(data))
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			mux := auth_api.ServeMux()
			mux.ServeHTTP(rr, req)
			if status := rr.Code; status != test.status {
				t.Fatalf("test failed: %v", status)
			}
		})
	}
}

func TestToken(t *testing.T) {
	tests := []struct {
		user   auth.User
		status int
		token  string
	}{
		{auth.User{Username: "user", Password: "password"}, http.StatusOK,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{auth.User{Username: "guest", Password: "password"}, http.StatusOK,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{auth.User{Username: "user", Password: "wrong"}, http.StatusUnauthorized, ""},
		{auth.User{Username: "johndoe", Password: "password"}, http.StatusNotFound, ""},
		{auth.User{Username: "username"}, http.StatusBadRequest, ""},
		{auth.User{Password: "password"}, http.StatusBadRequest, ""},
		{auth.User{}, http.StatusBadRequest, ""},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(test.user)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest("POST", "/token",
				bytes.NewBuffer(data))
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			mux := auth_api.ServeMux()
			mux.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			token := string(body)
			if token != test.token {
				t.Fatalf("test failed: %v", token)
			}
		})
	}
}
