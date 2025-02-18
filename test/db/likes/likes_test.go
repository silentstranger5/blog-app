package likes_test

import (
	"blog/config"
	"blog/db/auth"
	"blog/db/likes"
	"blog/db/posts"
	"fmt"
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

func TestAddLike(t *testing.T) {
	tests := []struct {
		userid, postid, count int
		ltype                 string
		error                 bool
	}{
		{1, 1, 1, "like", false},
		{1, 1, 0, "like", false},
		{1, 1, 1, "like", false},
		{1, 1, -1, "dislike", false},
		{1, 1, 0, "dislike", false},
		{1, 1, -1, "dislike", false},
		{1, 1, 1, "like", false},
		{2, 1, 2, "like", false},
		{2, 1, 1, "like", false},
		{2, 1, 2, "like", false},
		{2, 1, 0, "dislike", false},
		{2, 1, 1, "dislike", false},
		{2, 1, 0, "dislike", false},
		{2, 1, 2, "like", false},
		{1, 1, 1, "invalid", true},
		{1, 3, 0, "like", true},
		{3, 1, 0, "like", true},
		{3, 3, 0, "like", true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := likes.AddLike(config.DB, config.Ctx,
				test.userid, test.postid, test.ltype)
			if (err != nil) != test.error {
				t.Fatalf("test failed: %v", err)
			}
			if err != nil {
				return
			}
			count, err := likes.GetLikes(config.DB, config.Ctx, test.postid)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if count != test.count {
				t.Fatalf("test failed: %v", count)
			}
		})
	}
}

func TestGetLikes(t *testing.T) {
	tests := []struct {
		postid, count int
	}{
		{1, 2},
		{2, 0},
		{3, 0},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			count, err := likes.GetLikes(config.DB, config.Ctx, test.postid)
			if err != nil {
				t.Fatalf("test failed: %v", err)
			}
			if count != test.count {
				t.Fatalf("test failed: %v", count)
			}
		})
	}
}
