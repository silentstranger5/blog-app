package posts_test

import (
	"blog/config"
	"blog/db/auth"
	"blog/db/posts"
	"fmt"
	"os"
	"reflect"
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
		post  posts.Post
		error bool
	}{
		{posts.Post{AuthorId: 1, Title: "First Post", Text: "Hello, World!"}, false},
		{posts.Post{AuthorId: 1, Title: "Second Post", Text: "Your text here!"}, false},
		{posts.Post{AuthorId: 2, Title: "Third Post", Text: "Another post"}, false},
		{posts.Post{AuthorId: 1, Title: "New Post"}, true},
		{posts.Post{AuthorId: 1, Text: "Hello, World!"}, true},
		{posts.Post{AuthorId: 1}, true},
		{posts.Post{}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			_, err := posts.AddPost(config.DB, config.Ctx, test.post)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
		})
	}
}

func TestGetPosts(t *testing.T) {
	postList := []posts.Post{
		{Id: 1, AuthorId: 1, Author: "user", Title: "First Post", Text: "Hello, World!"},
		{Id: 2, AuthorId: 1, Author: "user", Title: "Second Post", Text: "Your text here!"},
		{Id: 3, AuthorId: 2, Author: "guest", Title: "Third Post", Text: "Another post"},
	}
	dbPostList, err := posts.GetPosts(config.DB, config.Ctx)
	if err != nil {
		t.Fatalf("test failed: %v", err)
	}
	var zero time.Time
	for i := range dbPostList {
		dbPostList[i].Created = zero
	}
	if !reflect.DeepEqual(postList, dbPostList) {
		t.Fatalf("test failed: %v", dbPostList)
	}
}

func TestGetPost(t *testing.T) {
	tests := []struct {
		post  posts.Post
		error bool
	}{
		{posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "First Post", Text: "Hello, World!"}, false},
		{posts.Post{Id: 2, AuthorId: 1, Author: "user", Title: "Second Post", Text: "Your text here!"}, false},
		{posts.Post{Id: 3, AuthorId: 2, Author: "guest", Title: "Third Post", Text: "Another post"}, false},
		{posts.Post{Id: 4}, true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			post, err := posts.GetPost(config.DB, config.Ctx, test.post.Id)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
			if err != nil {
				return
			}
			var zero time.Time
			post.Created = zero
			if !reflect.DeepEqual(post, test.post) {
				t.Fatalf("test failed, %v", err)
			}
		})
	}
}

func TestUpdatePost(t *testing.T) {
	tests := []struct {
		post posts.Post
	}{
		{posts.Post{Id: 1, AuthorId: 1, Author: "user", Title: "New First Post", Text: "New Text"}},
		{posts.Post{Id: 2, AuthorId: 1, Author: "user", Title: "New Second Post", Text: "New Text"}},
		{posts.Post{Id: 3, AuthorId: 2, Author: "guest", Title: "New Third Post", Text: "New Text"}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := posts.UpdatePost(config.DB, config.Ctx, test.post.Id, test.post)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			post, err := posts.GetPost(config.DB, config.Ctx, test.post.Id)
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

func TestFilterQuery(t *testing.T) {
	tests := []struct {
		query string
		posts []posts.Post
	}{
		{"first", []posts.Post{{Id: 1, AuthorId: 1, Author: "user", Title: "New First Post", Text: "New Text"}}},
		{"post", []posts.Post{
			{Id: 1, AuthorId: 1, Author: "user", Title: "New First Post", Text: "New Text"},
			{Id: 2, AuthorId: 1, Author: "user", Title: "New Second Post", Text: "New Text"},
			{Id: 3, AuthorId: 2, Author: "guest", Title: "New Third Post", Text: "New Text"},
		}},
		{"query", nil},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			postlist, err := posts.FilterQuery(config.DB, config.Ctx, test.query)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			var zero time.Time
			for i := range postlist {
				postlist[i].Created = zero
			}
			if !reflect.DeepEqual(postlist, test.posts) {
				t.Fatalf("test failed: %v", postlist)
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	postids := []int{1, 2, 3}
	for id := range postids {
		err := posts.DeletePost(config.DB, config.Ctx, id)
		if err != nil {
			t.Fatalf("test failed: %v", err)
		}
		_, err = posts.GetPost(config.DB, config.Ctx, id)
		if err == nil {
			t.Fatalf("test failed: %v", id)
		}
	}
}
