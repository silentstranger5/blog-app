package tags_test

import (
	posts_api "blog/api/posts"
	tags_api "blog/api/tags"
	"blog/config"
	"blog/db/auth"
	"blog/db/posts"
	"blog/db/tags"
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
	m.Run()
}

func TestAddTags(t *testing.T) {
	tests := []struct {
		tags   []tags.Tag
		token  string
		status int
	}{
		{[]tags.Tag{{Name: "first"}, {Name: "second"}, {Name: "third"}},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{[]tags.Tag{{Name: "second"}, {Name: "third"}},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{[]tags.Tag{{Name: "first tag"}, {Name: "second tag"}, {Name: "third tag"}},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK},
		{[]tags.Tag{{Name: "tag"}, {Name: "tag"}, {Name: "tag"}},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusOK},
		{[]tags.Tag{{Name: "tag"}},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusOK},
		{[]tags.Tag{{Name: "tag"}, {}},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusBadRequest},
		{nil, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusOK},
		{[]tags.Tag{{Name: "tag"}}, "", http.StatusUnauthorized},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			post := posts.Post{
				Title: "Title", Text: "Text", Tags: test.tags,
			}
			data, err := json.Marshal(post)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest(
				"POST", "/add", bytes.NewBuffer(data),
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(posts_api.AddPost)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	tests := []struct {
		postid, status int
		taglist        []tags.Tag
	}{
		{1, http.StatusOK, []tags.Tag{{Name: "first"}, {Name: "second"}, {Name: "third"}}},
		{2, http.StatusOK, []tags.Tag{{Name: "second"}, {Name: "third"}}},
		{3, http.StatusOK, []tags.Tag{{Name: "first tag"}, {Name: "second tag"}, {Name: "third tag"}}},
		{4, http.StatusOK, []tags.Tag{{Name: "tag"}}},
		{5, http.StatusOK, []tags.Tag{{Name: "tag"}}},
		{6, http.StatusNotFound, nil},
		{7, http.StatusNotFound, nil},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest(
				"GET", "/get/"+strconv.Itoa(test.postid), nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(tags_api.GetTags)
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
			var taglist []tags.Tag
			err = json.Unmarshal(body, &taglist)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if !reflect.DeepEqual(taglist, test.taglist) {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetPosts(t *testing.T) {
	tests := []struct {
		tag     string
		postids []int
		status  int
	}{
		{"first", []int{1}, http.StatusOK},
		{"second", []int{1, 2}, http.StatusOK},
		{"third", []int{1, 2}, http.StatusOK},
		{"tag", []int{4, 5}, http.StatusOK},
		{"empty", nil, http.StatusNotFound},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest(
				"GET", "/post/"+test.tag, nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(tags_api.GetPosts)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", err)
			}
			if status != http.StatusOK {
				return
			}
			body, err := io.ReadAll(rr.Body)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var postlist []posts.Post
			err = json.Unmarshal(body, &postlist)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var postids []int
			for _, post := range postlist {
				postids = append(postids, post.Id)
			}
			if !reflect.DeepEqual(postids, test.postids) {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}
