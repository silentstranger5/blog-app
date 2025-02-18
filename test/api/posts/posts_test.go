package posts_test

import (
	posts_api "blog/api/posts"
	"blog/config"
	"blog/db/auth"
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
		{Id: 0, Username: "user", Password: "password"},
		{Id: 1, Username: "guest", Password: "password"},
	}
	for _, user := range users {
		err = auth.AddUser(config.DB, config.Ctx, user)
		if err != nil {
			panic(err)
		}
	}
	m.Run()
}

func TestAddPost(t *testing.T) {
	tests := []struct {
		post   posts.Post
		status int
		token  string
	}{
		{posts.Post{Title: "New Post", Text: "Hello, World!"}, http.StatusOK,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{posts.Post{Title: "New Post", Text: "Hello, World!"}, http.StatusOK,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{posts.Post{Title: "Another Post", Text: "Your text here!"}, http.StatusOK,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{posts.Post{Title: "New Post"}, http.StatusBadRequest,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{posts.Post{Text: "Hello, World!"}, http.StatusBadRequest,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{posts.Post{}, http.StatusBadRequest,
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{posts.Post{}, http.StatusUnauthorized,
			""},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(test.post)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest("POST", "/add",
				bytes.NewBuffer(data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+test.token)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(posts_api.AddPost)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
		})
	}
}

func TestGetPosts(t *testing.T) {
	postList := []posts.Post{
		{Id: 1, AuthorId: 1, Title: "New Post", Author: "user", Text: "Hello, World!"},
		{Id: 2, AuthorId: 1, Title: "New Post", Author: "user", Text: "Hello, World!"},
		{Id: 3, AuthorId: 2, Title: "Another Post", Author: "guest", Text: "Your text here!"},
	}
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(posts_api.GetPosts)
	handler.ServeHTTP(rr, req)
	status := rr.Code
	if status != http.StatusOK {
		t.Fatalf("test failed: %v", status)
	}
	body, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
	var apiPostList []posts.Post
	err = json.Unmarshal(body, &apiPostList)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
	var zero time.Time
	for i := range apiPostList {
		apiPostList[i].Created = zero
	}
	if !reflect.DeepEqual(postList, apiPostList) {
		t.Fatalf("test failed: %v", apiPostList)
	}
}

func TestGetPost(t *testing.T) {
	tests := []struct {
		post   posts.Post
		status int
	}{
		{posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "New Post", Text: "Hello, World!"},
			http.StatusOK},
		{posts.Post{Id: 2, AuthorId: 1, Author: "user", Title: "New Post", Text: "Hello, World!"},
			http.StatusOK},
		{posts.Post{Id: 3, AuthorId: 2, Author: "guest", Title: "Another Post", Text: "Your text here!"},
			http.StatusOK},
		{posts.Post{Id: 4}, http.StatusNotFound},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest("GET",
				"/get/"+strconv.Itoa(test.post.Id), nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(posts_api.GetPost)
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
			var post posts.Post
			err = json.Unmarshal(body, &post)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			post.Created = zero
			if !reflect.DeepEqual(post, test.post) {
				t.Fatalf("test failed: %v", post)
			}
		})
	}
}

func TestUpdatePost(t *testing.T) {
	tests := []struct {
		post   posts.Post
		token  string
		status int
	}{
		{
			posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "First Post", Text: "New Text"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK,
		},
		{
			posts.Post{Id: 2, AuthorId: 1, Author: "user", Title: "Second Post", Text: "New Text"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusOK,
		},
		{
			posts.Post{Id: 3, AuthorId: 2, Author: "guest", Title: "New Post", Text: "Another Post"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusOK,
		},
		{
			posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "New Title", Text: "New Text"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA",
			http.StatusForbidden,
		},
		{
			posts.Post{Id: 3, AuthorId: 1, Author: "user", Title: "New Title", Text: "New Text"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusForbidden,
		},
		{
			posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "New Post"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusBadRequest,
		},
		{
			posts.Post{Id: 1, AuthorId: 1, Author: "user", Text: "Hello, World!"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusBadRequest,
		},
		{
			posts.Post{Id: 1, AuthorId: 1, Author: "user"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusBadRequest,
		},
		{
			posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "New Title", Text: "New Text"},
			"", http.StatusUnauthorized,
		},
		{
			posts.Post{Id: 4, AuthorId: 1, Author: "user", Title: "New Title", Text: "New Text"},
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I",
			http.StatusNotFound,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			data, err := json.Marshal(test.post)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req, err := http.NewRequest(
				"PUT", "/update/"+strconv.Itoa(test.post.Id),
				bytes.NewBuffer(data))
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+test.token)
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(posts_api.UpdatePost)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			req, err = http.NewRequest("GET",
				"/get/"+strconv.Itoa(test.post.Id), nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			handler = http.HandlerFunc(posts_api.GetPost)
			handler.ServeHTTP(rr, req)
			status = rr.Code
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
			var post posts.Post
			err = json.Unmarshal(body, &post)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			post.Created = zero
			if !reflect.DeepEqual(post, test.post) {
				t.Fatalf("test failed: %v", post)
			}
		})
	}
}

func TestSearchPost(t *testing.T) {
	tests := []struct {
		query   string
		postids []int
		status  int
	}{
		{"first", []int{1}, http.StatusOK},
		{"second", []int{2}, http.StatusOK},
		{"new", []int{3}, http.StatusOK},
		{"post", []int{1, 2, 3}, http.StatusOK},
		{"query", nil, http.StatusNotFound},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			req, err := http.NewRequest(
				"GET", "/search/"+test.query, nil,
			)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(posts_api.SearchPost)
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
				t.Fatalf("test failed: %v", postids)
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	tests := []struct {
		postid, status int
		token          string
	}{
		{1, http.StatusUnauthorized, ""},
		{2, http.StatusForbidden, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{3, http.StatusForbidden, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{1, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{2, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxfQ.jYyRJbb0WImFoUUdcslQQfwnXTHJzne-6tsPd8Hrw0I"},
		{3, http.StatusOK, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
		{4, http.StatusNotFound, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyfQ.9YCOE7tXJFvXEkLKezdd42NArXH6JXLtHbQu-KrwQSA"},
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
			handler := http.HandlerFunc(posts_api.DeletePost)
			handler.ServeHTTP(rr, req)
			status := rr.Code
			if status != test.status {
				t.Fatalf("test failed: %v", status)
			}
			if status != http.StatusOK {
				return
			}
			req, err = http.NewRequest("GET",
				"/get/"+strconv.Itoa(test.postid), nil)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			rr = httptest.NewRecorder()
			handler = http.HandlerFunc(posts_api.GetPost)
			handler.ServeHTTP(rr, req)
			status = rr.Code
			if status != http.StatusNotFound {
				t.Fatalf("test failed: %v", status)
			}
		})
	}
}
