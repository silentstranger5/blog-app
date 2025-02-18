package comments_test

import (
	comments_api "blog/api/comments"
	"blog/config"
	"blog/db/auth"
	"blog/db/comments"
	"blog/db/posts"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
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

func TestAddComment(t *testing.T) {
	tests := []struct {
		comment comments.Comment
		token   string
		status  int
	}{
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "First Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{comments.Comment{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "Second Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{comments.Comment{Id: 3, AuthorId: 1, PostId: 2, Author: "user", Text: "Third Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{comments.Comment{Id: 4, AuthorId: 2, PostId: 2, Author: "guest", Text: "Fourth Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusOK},
		{comments.Comment{Id: 5, AuthorId: 1, PostId: 1, Author: "user"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusBadRequest},
		{comments.Comment{Id: 5, AuthorId: 1, Author: "user"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusBadRequest},
		{comments.Comment{Id: 5, AuthorId: 1, PostId: 1, Author: "user", Text: "New Comment"}, "",
			http.StatusUnauthorized},
		{comments.Comment{Id: 5, AuthorId: 1, PostId: 3, Author: "user", Text: "New Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusNotFound},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(test.comment)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest(
				"POST", "/add/"+strconv.Itoa(test.comment.PostId),
				bytes.NewBuffer(data),
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(comments_api.AddComment)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			req, err = http.NewRequest(
				"GET", "/get/"+strconv.Itoa(test.comment.Id), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			handler = http.HandlerFunc(comments_api.GetComment)
			handler.ServeHTTP(rr, req)
			status = rr.Code
			if status != http.StatusOK {
				t.Fatalf("test failed: %v", status)
			}
			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var comment comments.Comment
			err = json.Unmarshal(body, &comment)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			comment.Created = zero
			if comment != test.comment {
				t.Fatalf("test failed: %v", comment)
			}
		})
	}
}

func TestGetComments(t *testing.T) {
	tests := []struct {
		postid, status int
		commentlist    []comments.Comment
	}{
		{1, http.StatusOK, []comments.Comment{
			{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "First Comment"},
			{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "Second Comment"},
		}},
		{2, http.StatusOK, []comments.Comment{
			{Id: 3, AuthorId: 1, PostId: 2, Author: "user", Text: "Third Comment"},
			{Id: 4, AuthorId: 2, PostId: 2, Author: "guest", Text: "Fourth Comment"},
		}},
		{3, http.StatusNotFound, nil},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest(
				"GET", "/post/"+strconv.Itoa(test.postid), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(comments_api.GetComments)
			handler.ServeHTTP(rr, req)
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
			var commentlist []comments.Comment
			err = json.Unmarshal(body, &commentlist)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			for i := range commentlist {
				commentlist[i].Created = zero
			}
			if !reflect.DeepEqual(commentlist, test.commentlist) {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetComment(t *testing.T) {
	tests := []struct {
		comment comments.Comment
		status  int
	}{
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "First Comment"}, http.StatusOK},
		{comments.Comment{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "Second Comment"}, http.StatusOK},
		{comments.Comment{Id: 3, AuthorId: 1, PostId: 2, Author: "user", Text: "Third Comment"}, http.StatusOK},
		{comments.Comment{Id: 4, AuthorId: 2, PostId: 2, Author: "guest", Text: "Fourth Comment"}, http.StatusOK},
		{comments.Comment{Id: 5}, http.StatusNotFound},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest(
				"GET", "/get/"+strconv.Itoa(test.comment.Id), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(comments_api.GetComment)
			handler.ServeHTTP(rr, req)
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
			var comment comments.Comment
			err = json.Unmarshal(body, &comment)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			comment.Created = zero
			if comment != test.comment {
				t.Fatalf("test failed: %v", comment)
			}
		})
	}
}

func TestUpdateComment(t *testing.T) {
	tests := []struct {
		comment comments.Comment
		token   string
		status  int
	}{
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "New First Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{comments.Comment{Id: 2, AuthorId: 1, PostId: 1, Author: "user", Text: "New Second Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{comments.Comment{Id: 3, AuthorId: 1, PostId: 2, Author: "user", Text: "New Third Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{comments.Comment{Id: 4, AuthorId: 2, PostId: 2, Author: "guest", Text: "New Fourth Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusOK},
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "First Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusForbidden},
		{comments.Comment{Id: 4, AuthorId: 2, PostId: 1, Author: "guest", Text: "New Third Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusForbidden},
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user", Text: "New Comment"},
			"",
			http.StatusUnauthorized},
		{comments.Comment{Id: 1, AuthorId: 1, PostId: 1, Author: "user"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusBadRequest},
		{comments.Comment{Id: 1, AuthorId: 1, Author: "user"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusBadRequest},
		{comments.Comment{Id: 5, AuthorId: 1, Author: "user", Text: "New Comment"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusNotFound},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(test.comment)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest(
				"POST", "/update/"+strconv.Itoa(test.comment.Id),
				bytes.NewBuffer(data),
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(comments_api.UpdateComment)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			req, err = http.NewRequest(
				"GET", "/get/"+strconv.Itoa(test.comment.Id), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			handler = http.HandlerFunc(comments_api.GetComment)
			handler.ServeHTTP(rr, req)
			status = rr.Code
			if status != http.StatusOK {
				t.Fatalf("test failed: %v", status)
			}
			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var comment comments.Comment
			err = json.Unmarshal(body, &comment)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			comment.Created = zero
			if comment != test.comment {
				t.Fatalf("test failed: %v", comment)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	tests := []struct {
		postid, status int
		token          string
	}{
		{1, http.StatusUnauthorized, ""},
		{1, http.StatusForbidden, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{4, http.StatusForbidden, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{2, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{3, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{4, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{5, http.StatusNotFound, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest(
				"DELETE", "/delete/"+strconv.Itoa(test.postid), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+test.token)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(comments_api.DeleteComment)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			req, err = http.NewRequest(
				"GET", "/get/"+strconv.Itoa(test.postid), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			handler = http.HandlerFunc(comments_api.GetComment)
			handler.ServeHTTP(rr, req)
			status = rr.Code
			if status != http.StatusNotFound {
				t.Fatalf("test failed: %v", status)
			}
		})
	}
}
