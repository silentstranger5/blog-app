package likes_test

import (
	posts_api "blog/api/posts"
	"blog/config"
	"blog/db/auth"
	"blog/db/posts"
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
	users := []auth.User{
		{Id: 1, Username: "user", Password: "password"},
		{Id: 2, Username: "guest", Password: "password"},
	}
	for _, user := range users {
		err = auth.AddUser(config.DB, config.Ctx, user)
		if err != nil {
			panic(err)
		}
	}
	postList := []posts.Post{
		{AuthorId: 1, Title: "New Post", Text: "Post Text"},
		{AuthorId: 2, Title: "New Post", Text: "Post Text"},
	}
	for _, post := range postList {
		_, err = posts.AddPost(config.DB, config.Ctx, post)
		if err != nil {
			panic(err)
		}
	}
	m.Run()
}

func TestLikePost(t *testing.T) {
	tests := []struct {
		postid, status, count int
		token                 string
	}{
		{1, http.StatusOK, 1, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, 0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, 1, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, 2, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{1, http.StatusOK, 1, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{1, http.StatusOK, 2, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{2, http.StatusOK, 1, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{2, http.StatusOK, 2, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{3, http.StatusNotFound, 0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusUnauthorized, 2, ""},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			url := fmt.Sprintf("/%d/like", test.postid)
			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+test.token)
			rr := httptest.NewRecorder()
			mux := posts_api.ServeMux()
			mux.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			url = fmt.Sprintf("/%d/likes", test.postid)
			req, err = http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			status = rr.Code
			if status != http.StatusOK {
				t.Fatalf("test failed: %v", err)
			}
			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var count int
			err = json.Unmarshal(body, &count)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if count != test.count {
				t.Fatalf("test failed: %v", count)
			}
		})
	}
}

func TestDislikePost(t *testing.T) {
	tests := []struct {
		postid, status, count int
		token                 string
	}{
		{1, http.StatusOK, 0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, 1, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, 0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, -2, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{1, http.StatusOK, -1, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{1, http.StatusOK, -2, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{2, http.StatusOK, 0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{2, http.StatusOK, -2, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{3, http.StatusNotFound, 0, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{1, http.StatusUnauthorized, 0, ""},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			url := fmt.Sprintf("/%d/dislike", test.postid)
			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+test.token)
			rr := httptest.NewRecorder()
			mux := posts_api.ServeMux()
			mux.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", err)
			}
			if status != http.StatusOK {
				return
			}
			url = fmt.Sprintf("/%d/likes", test.postid)
			req, err = http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			mux.ServeHTTP(rr, req)
			status = rr.Code
			if status != http.StatusOK {
				t.Fatalf("test failed: %v", err)
			}
			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var count int
			err = json.Unmarshal(body, &count)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if count != test.count {
				t.Fatalf("test failed: %v", count)
			}
		})
	}
}
